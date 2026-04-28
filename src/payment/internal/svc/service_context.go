package svc

import (
	"payment/ent"
	"payment/internal/config"
	"payment/internal/provider"

	"shared/auth"
	sharedevent "shared/event"
)

type ServiceContext struct {
	Config          config.Config
	Db              *ent.Client
	Dispatcher      sharedevent.Dispatcher[sharedevent.Event]
	PaymentProvider provider.PaymentProvider
	Keycloak        auth.KeycloakConfig
	Validator       auth.TokenValidator
	ServiceClient   *auth.ServiceClient
}

func NewServiceContext(c config.Config, db *ent.Client, dispatcher sharedevent.Dispatcher[sharedevent.Event]) *ServiceContext {
	return &ServiceContext{
		Config:          c,
		Db:              db,
		Dispatcher:      dispatcher,
		PaymentProvider: provider.NewMockProvider(),
		Keycloak:        c.Keycloak,
		Validator:       auth.NewKeycloakValidator(c.Keycloak),
		ServiceClient:   auth.NewServiceClient(auth.ServiceClientConfig{KeycloakConfig: c.Keycloak}),
	}
}
