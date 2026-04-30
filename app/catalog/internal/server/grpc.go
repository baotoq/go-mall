package server

import (
	"context"

	v1 "gomall/api/catalog/v1"
	"gomall/app/catalog/internal/conf"
	"gomall/app/catalog/internal/service"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func NewGRPCServer(c *conf.Server, catalog *service.CatalogService, logger log.Logger) *grpc.Server {
	jwks, err := keyfunc.NewDefault([]string{c.GetAuth().GetJwksUrl()})
	if err != nil {
		panic(err)
	}
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			selector.Server(
				jwt.Server(jwks.Keyfunc, jwt.WithSigningMethod(jwtv5.SigningMethodRS256)),
			).Match(func(_ context.Context, op string) bool {
				_, public := publicRoutes[op]
				return !public
			}).Build(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterCatalogServiceServer(srv, catalog)
	return srv
}
