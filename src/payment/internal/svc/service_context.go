package svc

import (
	"payment/ent"
	"payment/internal/config"
	"payment/internal/provider"

	sharedevent "shared/event"
)

type ServiceContext struct {
	Config          config.Config
	Db              *ent.Client
	Dispatcher      sharedevent.Dispatcher[sharedevent.Event]
	PaymentProvider provider.PaymentProvider
}

func NewServiceContext(c config.Config, db *ent.Client, dispatcher sharedevent.Dispatcher[sharedevent.Event]) *ServiceContext {
	return &ServiceContext{
		Config:          c,
		Db:              db,
		Dispatcher:      dispatcher,
		PaymentProvider: provider.NewMockProvider(),
	}
}
