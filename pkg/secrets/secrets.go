package secrets

const (
	keyDatabase = "DATABASE_CONNECTION_STRING"
	keyKeycloak = "KEYCLOAK_JWKS_URL"
)

type Secrets struct {
	DatabaseConnectionString string
	KeycloakJWKSURL          string
}

func Parse(m map[string]string, serviceDBKey string) Secrets {
	dsn := m[serviceDBKey]
	if dsn == "" {
		dsn = m[keyDatabase]
	}
	return Secrets{
		DatabaseConnectionString: dsn,
		KeycloakJWKSURL:          m[keyKeycloak],
	}
}
