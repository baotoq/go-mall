// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"payment/ent"
	"payment/internal/config"
	"payment/internal/event"
	"payment/internal/provider"
)

type ServiceContext struct {
	Config          config.Config
	Db              *ent.Client
	Dispatcher      event.Dispatcher[event.Event]
	PaymentProvider provider.PaymentProvider
}

func NewServiceContext(c config.Config, db *ent.Client, dispatcher event.Dispatcher[event.Event]) *ServiceContext {
	return &ServiceContext{
		Config:          c,
		Db:              db,
		Dispatcher:      dispatcher,
		PaymentProvider: provider.NewMockProvider(),
	}
}
