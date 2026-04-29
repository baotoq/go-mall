package config

import (
	"github.com/zeromicro/go-zero/rest"

	"shared/auth"
)

type Config struct {
	rest.RestConf
	Keycloak     auth.KeycloakConfig
	DB           struct{ DSN string }
	CatalogAppID string
	PaymentAppID string
}
