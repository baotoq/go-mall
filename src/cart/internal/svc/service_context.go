package svc

import (
	"cart/ent"
	"cart/internal/clients/catalog"
	"cart/internal/clients/payment"
	"cart/internal/config"

	sharedevent "shared/event"
)

type ServiceContext struct {
	Config         config.Config
	Db             *ent.Client
	Dispatcher     sharedevent.Dispatcher[sharedevent.Event]
	CatalogClient  *catalog.Client
	PaymentClient  *payment.Client
}

func NewServiceContext(
	c config.Config,
	db *ent.Client,
	dispatcher sharedevent.Dispatcher[sharedevent.Event],
	catalogClient *catalog.Client,
	paymentClient *payment.Client,
) *ServiceContext {
	return &ServiceContext{
		Config:        c,
		Db:            db,
		Dispatcher:    dispatcher,
		CatalogClient: catalogClient,
		PaymentClient: paymentClient,
	}
}
