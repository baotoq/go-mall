package secrets

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-kratos/kratos/v2/log"
)

// SecretStore is the interface for retrieving secrets from a store.
type SecretStore interface {
	GetSecret(ctx context.Context, storeName, key string, meta map[string]string) (map[string]string, error)
}

type options struct {
	storeName   string
	key         string
	maxAttempts int
	interval    time.Duration
}

type Option func(*options)

func WithStoreName(name string) Option {
	return func(o *options) { o.storeName = name }
}

func WithKey(key string) Option {
	return func(o *options) { o.key = key }
}

func WithMaxAttempts(n int) Option {
	return func(o *options) {
		if n > 0 {
			o.maxAttempts = n
		}
	}
}

func WithInterval(d time.Duration) Option {
	return func(o *options) { o.interval = d }
}

// LoadSecrets retries GetSecret until the secret store is ready and returns
// the raw secret map. Defaults: storeName="secretstore", key="secrets",
// maxAttempts=5. Returns an error when exhausted.
func LoadSecrets(dc SecretStore, logger log.Logger, opts ...Option) (map[string]string, error) {
	o := &options{
		storeName:   "secretstore",
		key:         "secrets",
		maxAttempts: 5,
		interval:    5 * time.Second,
	}
	for _, opt := range opts {
		opt(o)
	}

	helper := log.NewHelper(logger)
	var secret map[string]string

	err := backoff.RetryNotify(
		func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			s, err := dc.GetSecret(ctx, o.storeName, o.key, nil)
			if err != nil {
				return err
			}
			secret = s
			return nil
		},
		backoff.WithMaxRetries(backoff.NewConstantBackOff(o.interval), uint64(o.maxAttempts-1)),
		func(err error, d time.Duration) {
			helper.Infof("waiting for secret store: %v (retry in %s)", err, d)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("secret store not ready after %ds: %w", o.maxAttempts*5, err)
	}
	return secret, nil
}
