// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"product/ent"
	"product/internal/config"
)

type ServiceContext struct {
	Config config.Config
	Db     *ent.Client
}

func NewServiceContext(c config.Config, db *ent.Client) *ServiceContext {
	return &ServiceContext{
		Config: c,
		Db:     db,
	}
}
