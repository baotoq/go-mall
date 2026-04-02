// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"catalog/ent"
	"catalog/internal/config"
	"catalog/internal/event"
)

type ServiceContext struct {
	Config     config.Config
	Db         *ent.Client
	Dispatcher event.Dispatcher[event.Event]
}

func NewServiceContext(c config.Config, db *ent.Client, dispatcher event.Dispatcher[event.Event]) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		Db:         db,
		Dispatcher: dispatcher,
	}
}
