// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/zeromicro/go-zero/rest"

	"shared/auth"
)

type Config struct {
	rest.RestConf
	Keycloak auth.KeycloakConfig
	DB       struct{ DSN string }
}
