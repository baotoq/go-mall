package outbox

import (
	"context"
	"database/sql"

	"github.com/go-kratos/kratos/v2/log"
)

func NewWithPublisher(db *sql.DB, pub publisher, cfg Config, logger log.Logger) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Client{
		db:     db,
		pub:    pub,
		cfg:    cfg,
		log:    log.NewHelper(logger),
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}, nil
}

func (c *Client) ProcessBatch(ctx context.Context) error {
	return c.processBatch(ctx)
}

func (c *Client) RunSweep(ctx context.Context) {
	retentionSecs := c.cfg.RetentionPeriod.Seconds()
	_, _ = c.db.ExecContext(ctx, sweepOutboxSQL, retentionSecs)
	_, _ = c.db.ExecContext(ctx, sweepInboxSQL, retentionSecs)
}
