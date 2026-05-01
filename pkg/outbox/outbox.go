package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

// publisher is a private interface for testability — satisfied by dapr.Client.
type publisher interface {
	PublishEvent(ctx context.Context, pubsubName, topicName string, data interface{}, opts ...dapr.PublishEventOption) error
}

// TxExecer is the minimal SQL interface satisfied by *sql.Tx.
type TxExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type publishOptions struct {
	headers   map[string]string
	messageID string
}

// PublishOption configures a Publish call.
type PublishOption func(*publishOptions)

// WithHeaders attaches metadata headers to the outbox message.
func WithHeaders(h map[string]string) PublishOption {
	return func(o *publishOptions) { o.headers = h }
}

// WithMessageID overrides the auto-generated UUID message ID.
func WithMessageID(id string) PublishOption {
	return func(o *publishOptions) { o.messageID = id }
}

// Client is the outbox/inbox coordinator.
type Client struct {
	db        *sql.DB
	pub       publisher
	cfg       Config
	log       *log.Helper
	stopCh    chan struct{}
	doneCh    chan struct{}
	stopOnce  sync.Once
	startOnce sync.Once
}

// New creates a new Client. Call Migrate before Start.
func New(db *sql.DB, daprClient dapr.Client, cfg Config, logger log.Logger) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Client{
		db:  db,
		pub: daprClient,
		cfg: cfg,
		log: log.NewHelper(logger),
	}, nil
}

// Migrate applies the outbox/inbox schema idempotently. Call once after DB schema creation.
func (c *Client) Migrate(ctx context.Context) error {
	return applySchema(ctx, c.db)
}

// Start launches the relay and sweeper goroutines. Idempotent.
func (c *Client) Start(ctx context.Context) error {
	if !c.cfg.EnableRelay {
		return nil
	}
	c.startOnce.Do(func() {
		c.stopCh = make(chan struct{})
		c.doneCh = make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); c.runRelay(ctx) }()
		go func() { defer wg.Done(); c.runSweeper(ctx) }()
		go func() { wg.Wait(); close(c.doneCh) }()
	})
	return nil
}

// Stop drains the relay and sweeper goroutines. Idempotent.
func (c *Client) Stop(ctx context.Context) error {
	if !c.cfg.EnableRelay {
		return nil
	}
	c.stopOnce.Do(func() {
		if c.stopCh != nil {
			close(c.stopCh)
		}
	})
	if c.doneCh != nil {
		select {
		case <-c.doneCh:
		case <-ctx.Done():
		}
	}
	return nil
}

// Publish enqueues a message in the outbox table within the provided transaction.
// The message is only delivered if tx commits. Returns the assigned message ID.
func (c *Client) Publish(ctx context.Context, tx TxExecer, topic string, payload any, opts ...PublishOption) (string, error) {
	po := &publishOptions{headers: make(map[string]string)}
	for _, o := range opts {
		o(po)
	}
	if po.messageID == "" {
		po.messageID = uuid.NewString()
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	headersBytes, err := json.Marshal(po.headers)
	if err != nil {
		return "", err
	}

	_, err = tx.ExecContext(ctx, insertOutboxSQL, po.messageID, topic, string(payloadBytes), string(headersBytes))
	if err != nil {
		return "", err
	}
	return po.messageID, nil
}
