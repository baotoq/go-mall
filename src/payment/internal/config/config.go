// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import (
	"github.com/zeromicro/go-zero/rest"

	"shared/auth"
)

type Config struct {
	rest.RestConf
	Keycloak auth.KeycloakConfig
}
