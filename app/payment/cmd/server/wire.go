//go:build wireinject
// +build wireinject

package main

import (
	"gomall/app/payment/internal/biz"
	"gomall/app/payment/internal/conf"
	"gomall/app/payment/internal/data"
	"gomall/app/payment/internal/server"
	"gomall/app/payment/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

func wireApp(*conf.Server, *conf.Data, *conf.Auth, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
