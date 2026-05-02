package server

import (
	v1 "gomall/api/payment/v1"
	"gomall/app/payment/internal/conf"
	"gomall/app/payment/internal/service"
	pkgserver "gomall/pkg/server"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	kratosJWT "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func NewHTTPServer(c *conf.Server, auth *conf.Auth, payment *service.PaymentService, sub *PaymentSubscriber, logger log.Logger) *http.Server {
	mw := []middleware.Middleware{recovery.Recovery()}
	if auth.JwksURL != "" {
		jwks, err := keyfunc.NewDefault([]string{auth.JwksURL})
		if err != nil {
			panic(err)
		}
		mw = append(mw, kratosJWT.Server(jwks.Keyfunc, kratosJWT.WithSigningMethod(jwtv5.SigningMethodRS256)))
	}
	var opts = []http.ServerOption{
		http.Middleware(mw...),
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
	sub.Register(srv)
	return srv
}
