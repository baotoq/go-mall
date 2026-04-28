package svc

import (
	"catalog/ent"
	"catalog/internal/config"

	sharedevent "shared/event"
)

type ServiceContext struct {
	Config     config.Config
	Db         *ent.Client
	Dispatcher sharedevent.Dispatcher[sharedevent.Event]
}

func NewServiceContext(c config.Config, db *ent.Client, dispatcher sharedevent.Dispatcher[sharedevent.Event]) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		Db:         db,
		Dispatcher: dispatcher,
	}
}
