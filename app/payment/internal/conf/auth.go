package conf

// Auth holds runtime auth configuration injected from the Dapr secret store.
// It is not a proto message to avoid requiring protoc regeneration.
type Auth struct {
	JwksURL string
}
