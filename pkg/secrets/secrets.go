package secrets

const (
	keyDatabase         = "DATABASE_CONNECTION_STRING"
	keyWorkflowDatabase = "WORKFLOWSTORE_DATABASE_CONNECTION_STRING"
	keyKeycloak         = "KEYCLOAK_JWKS_URL"
)

type Secrets struct {
	DatabaseConnectionString          string
	WorkflowstoreConnectionString     string
	KeycloakJWKSURL                   string
}

func Parse(m map[string]string, serviceDBKey string) Secrets {
	dsn := m[serviceDBKey]
	if dsn == "" {
		dsn = m[keyDatabase]
	}
	wdsn := m[keyWorkflowDatabase]
	if wdsn == "" {
		wdsn = m[keyDatabase]
	}
	return Secrets{
		DatabaseConnectionString:      dsn,
		WorkflowstoreConnectionString: wdsn,
		KeycloakJWKSURL:               m[keyKeycloak],
	}
}
