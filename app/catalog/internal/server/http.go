package server

import (
	"context"

	v1 "gomall/api/catalog/v1"
	"gomall/app/catalog/internal/conf"
	"gomall/app/catalog/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

var publicRoutes = map[string]struct{}{
	v1.OperationCatalogServiceListProducts:   {},
	v1.OperationCatalogServiceGetProduct:     {},
	v1.OperationCatalogServiceListCategories: {},
	v1.OperationCatalogServiceGetCategory:    {},
}

func NewHTTPServer(c *conf.Server, catalog *service.CatalogService, logger log.Logger) *http.Server {
	secret := []byte(c.GetAuth().GetSecret())
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			selector.Server(
				jwt.Server(func(_ *jwtv5.Token) (interface{}, error) {
					return secret, nil
				}),
			).Match(func(_ context.Context, op string) bool {
				_, public := publicRoutes[op]
				return !public
			}).Build(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterCatalogServiceHTTPServer(srv, catalog)
	return srv
}
