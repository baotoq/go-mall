package config

import (
	"github.com/zeromicro/go-zero/rest"

	"shared/auth"
)

type Config struct {
	rest.RestConf
	Keycloak     auth.KeycloakConfig
	CatalogAppID string
	PaymentAppID string
}
