package server

import (
	"os"

	v1 "gomall/api/payment/v1"
	"gomall/app/payment/internal/conf"
	"gomall/app/payment/internal/service"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	kratosJWT "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func NewGRPCServer(c *conf.Server, payment *service.PaymentService, logger log.Logger) *grpc.Server {
	mw := []middleware.Middleware{recovery.Recovery()}
	if jwksURL := os.Getenv("KEYCLOAK_JWKS_URL"); jwksURL != "" {
		jwks, err := keyfunc.NewDefault([]string{jwksURL})
		if err != nil {
			panic(err)
		}
		mw = append(mw, kratosJWT.Server(jwks.Keyfunc, kratosJWT.WithSigningMethod(jwtv5.SigningMethodRS256)))
	}
	var opts = []grpc.ServerOption{
		grpc.Middleware(mw...),
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
	v1.RegisterPaymentServiceServer(srv, payment)
	return srv
}
