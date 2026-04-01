// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"product/ent"
	"product/internal/config"
	"product/internal/event"

	dapr "github.com/dapr/go-sdk/client"
)

type ServiceContext struct {
	Config          config.Config
	Db              *ent.Client
	EventDispatcher event.EventDispatcher[event.Event]
}

func NewServiceContext(c config.Config, db *ent.Client, dapr dapr.Client) *ServiceContext {
	return &ServiceContext{
		Config:          c,
		Db:              db,
		EventDispatcher: event.NewDaprEventDispatcher[event.Event](dapr),
	}
}
