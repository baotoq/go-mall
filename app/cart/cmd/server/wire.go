//go:build wireinject
// +build wireinject

package main

import (
	"gomall/app/cart/internal/biz"
	"gomall/app/cart/internal/conf"
	"gomall/app/cart/internal/data"
	"gomall/app/cart/internal/server"
	"gomall/app/cart/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
