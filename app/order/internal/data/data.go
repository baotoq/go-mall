package data

import (
	"context"
	"database/sql"
	"os"
	"time"

	"gomall/app/order/internal/biz"
	"gomall/app/order/internal/conf"
	"gomall/app/order/internal/data/ent"
	"gomall/pkg/outbox"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	_ "github.com/lib/pq"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

var ProviderSet = wire.NewSet(
	ProvideSQLDB,
	ProvideOutboxConfig,
	outbox.ProviderSet,
	NewOutboxPublisher,
	NewData,
	NewOrderRepo,
)

type Data struct {
	db    *ent.Client
	sqlDB *sql.DB
}

func ProvideSQLDB(c *conf.Data) (*sql.DB, func(), error) {
	db, err := sql.Open(c.Database.Driver, c.Database.Source)
	if err != nil {
		return nil, nil, err
	}
	return db, func() { _ = db.Close() }, nil
}

func ProvideOutboxConfig() outbox.Config {
	cfg := outbox.DefaultConfig()
	cfg.EnableRelay = os.Getenv("OUTBOX_RELAY_ENABLED") != "false"
	return cfg
}

type outboxPub struct{ c *outbox.Client }

func (a *outboxPub) Publish(ctx context.Context, tx biz.TxExecer, topic string, payload any) (string, error) {
	return a.c.Publish(ctx, tx, topic, payload)
}

func NewOutboxPublisher(c *outbox.Client) biz.OutboxPublisher {
	return &outboxPub{c: c}
}

func NewData(sqlDB *sql.DB, ob *outbox.Client, logger log.Logger) (*Data, func(), error) {
	drv := entsql.OpenDB(dialect.Postgres, sqlDB)
	client := ent.NewClient(ent.Driver(drv))

	migrateCtx, migrateCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer migrateCancel()
	if err := client.Schema.Create(migrateCtx); err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	if err := ob.Migrate(migrateCtx); err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	cleanup := func() {
		if err := client.Close(); err != nil {
			log.NewHelper(logger).Error(err)
		}
	}
	return &Data{db: client, sqlDB: sqlDB}, cleanup, nil
}
