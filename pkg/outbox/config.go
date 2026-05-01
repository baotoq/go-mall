package outbox

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	PubsubName      string
	ConsumerID      string
	PollInterval    time.Duration
	BatchSize       int
	MaxAttempts     int
	RetentionPeriod time.Duration
	BackoffBase     time.Duration
	BackoffMax      time.Duration
	SweepInterval   time.Duration
	PublishTimeout  time.Duration
	EnableRelay     bool
}

func DefaultConfig() Config {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	return Config{
		PubsubName:      "pubsub",
		ConsumerID:      hostname,
		PollInterval:    5 * time.Second,
		BatchSize:       100,
		MaxAttempts:     5,
		RetentionPeriod: 24 * time.Hour,
		BackoffBase:     time.Second,
		BackoffMax:      5 * time.Minute,
		SweepInterval:   10 * time.Minute,
		PublishTimeout:  5 * time.Second,
		EnableRelay:     true,
	}
}

func (c *Config) Validate() error {
	if c.PubsubName == "" {
		return errors.New("outbox: PubsubName is required")
	}
	if c.BatchSize <= 0 {
		return errors.New("outbox: BatchSize must be positive")
	}
	if c.PollInterval <= 0 {
		return errors.New("outbox: PollInterval must be positive")
	}
	if c.MaxAttempts < 0 {
		return errors.New("outbox: MaxAttempts must be non-negative")
	}
	if c.BackoffBase <= 0 {
		return errors.New("outbox: BackoffBase must be positive")
	}
	if c.BackoffMax < c.BackoffBase {
		return errors.New("outbox: BackoffMax must be >= BackoffBase")
	}
	return nil
}
