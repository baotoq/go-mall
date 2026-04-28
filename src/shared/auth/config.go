package auth

import "time"

type KeycloakConfig struct {
	RealmURL     string
	ClientID     string
	ClientSecret string
	JWKSCacheTTL time.Duration
}