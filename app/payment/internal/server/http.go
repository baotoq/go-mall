package server

import (
	v1 "gomall/api/payment/v1"
	"gomall/app/payment/internal/conf"
	"gomall/app/payment/internal/service"
	pkgserver "gomall/pkg/server"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

func NewHTTPServer(c *conf.Server, payment *service.PaymentService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
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
	v1.RegisterPaymentServiceHTTPServer(srv, payment)
	srv.HandleFunc("/healthz", pkgserver.Healthz)
	return srv
}
