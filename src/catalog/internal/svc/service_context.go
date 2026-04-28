package svc

import (
	"catalog/ent"
	"catalog/internal/config"

	"shared/auth"
	sharedevent "shared/event"
)

type ServiceContext struct {
	Config        config.Config
	Db            *ent.Client
	Dispatcher    sharedevent.Dispatcher[sharedevent.Event]
	Keycloak      auth.KeycloakConfig
	Validator     auth.TokenValidator
	ServiceClient *auth.ServiceClient
}

func NewServiceContext(c config.Config, db *ent.Client, dispatcher sharedevent.Dispatcher[sharedevent.Event]) *ServiceContext {
	return &ServiceContext{
		Config:        c,
		Db:            db,
		Dispatcher:    dispatcher,
		Keycloak:      c.Keycloak,
		Validator:     auth.NewKeycloakValidator(c.Keycloak),
		ServiceClient: auth.NewServiceClient(auth.ServiceClientConfig{KeycloakConfig: c.Keycloak}),
	}
}
