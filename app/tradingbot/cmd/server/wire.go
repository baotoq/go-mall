//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"tradingbot/app/tradingbot/internal/biz"
	"tradingbot/app/tradingbot/internal/conf"
	"tradingbot/app/tradingbot/internal/data"
	"tradingbot/app/tradingbot/internal/server"
	"tradingbot/app/tradingbot/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
