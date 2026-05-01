package data

import (
	"context"
	"time"

	"gomall/app/payment/internal/conf"
	"gomall/app/payment/internal/data/ent"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	_ "github.com/lib/pq"
)

var ProviderSet = wire.NewSet(NewData, NewPaymentRepo)

type Data struct {
	db *ent.Client
}

func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	client, err := ent.Open(c.Database.Driver, c.Database.Source)
	if err != nil {
		return nil, nil, err
	}
	migrateCtx, migrateCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer migrateCancel()
	if err := client.Schema.Create(migrateCtx); err != nil {
		_ = client.Close()
		return nil, nil, err
	}
	cleanup := func() {
		if err := client.Close(); err != nil {
			log.NewHelper(logger).Error(err)
		}
	}
	return &Data{db: client}, cleanup, nil
}
