package server

import (
	v1 "gomall/api/cart/v1"
	"gomall/app/cart/internal/conf"
	"gomall/app/cart/internal/service"
	pkgserver "gomall/pkg/server"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

func NewHTTPServer(c *conf.Server, cart *service.CartService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Filter(pkgserver.CORSFilter),
		http.Middleware(
			recovery.Recovery(),
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
	v1.RegisterCartServiceHTTPServer(srv, cart)
	srv.HandleFunc("/healthz", pkgserver.Healthz)
	return srv
}
