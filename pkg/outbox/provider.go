package outbox

import (
	"context"
	"database/sql"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet provides the outbox Client for Wire injection.
// Callers must also provide: *sql.DB, dapr.Client, Config, log.Logger.
var ProviderSet = wire.NewSet(ProvideClient)

// ProvideClient constructs a Client and returns a cleanup that stops the relay.
// Migrate and Start are NOT called here — callers must invoke them at appropriate lifecycle points.
func ProvideClient(db *sql.DB, daprClient dapr.Client, cfg Config, logger log.Logger) (*Client, func(), error) {
	c, err := New(db, daprClient, cfg, logger)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = c.Stop(ctx)
	}
	return c, cleanup, nil
}
