//go:build wireinject
// +build wireinject

package main

import (
	"gomall/app/catalog/internal/biz"
	"gomall/app/catalog/internal/conf"
	"gomall/app/catalog/internal/data"
	"gomall/app/catalog/internal/server"
	"gomall/app/catalog/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
