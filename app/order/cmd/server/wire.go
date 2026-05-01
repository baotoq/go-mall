//go:build wireinject
// +build wireinject

package main

import (
	"gomall/app/order/internal/biz"
	"gomall/app/order/internal/conf"
	"gomall/app/order/internal/data"
	"gomall/app/order/internal/server"
	"gomall/app/order/internal/service"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

func wireApp(*conf.Server, *conf.Data, log.Logger, dapr.Client) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
