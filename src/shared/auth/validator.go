package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
)

var (
	ErrInvalidToken        = errors.New("invalid token")
	ErrTokenExpired        = errors.New("token expired")
	ErrKeycloakUnavailable = errors.New("keycloak unavailable")
)

type KeycloakValidator struct {
	config   KeycloakConfig
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	mu       sync.RWMutex
	initOnce sync.Once
	initErr  error
}

func NewKeycloakValidator(config KeycloakConfig) *KeycloakValidator {
	if config.JWKSCacheTTL <= 0 {
		config.JWKSCacheTTL = 24 * time.Hour
	}
	return &KeycloakValidator{config: config}
}

func (v *KeycloakValidator) Validate(ctx context.Context, rawToken string) (*Claims, error) {
	v.initOnce.Do(func() {
		v.initErr = v.lazyInit(ctx)
	})
	if v.initErr != nil {
		return nil, fmt.Errorf("%w: %v", ErrKeycloakUnavailable, v.initErr)
	}

	v.mu.RLock()
	verifier := v.verifier
	v.mu.RUnlock()

	idToken, err := verifier.Verify(ctx, rawToken)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("%w: %v", ErrKeycloakUnavailable, err)
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	var claims struct {
		RealmAccess struct {
			Roles []string `json:"roles"`
		} `json:"realm_access"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("%w: failed to parse claims: %v", ErrInvalidToken, err)
	}

	return &Claims{
		Subject: idToken.Subject,
		Roles:   claims.RealmAccess.Roles,
		Email:   claims.Email,
		Name:    claims.Name,
	}, nil
}

func (v *KeycloakValidator) lazyInit(ctx context.Context) error {
	issuer := v.config.RealmURL
	if issuer == "" {
		return errors.New("realm URL is empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return fmt.Errorf("oidc discovery failed: %w", err)
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	v.provider = provider
	v.verifier = provider.Verifier(&oidc.Config{
		ClientID:          v.config.ClientID,
		SkipClientIDCheck: true,
	})
	return nil
}

type ctxKey struct{}

func WithClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

func ClaimsFromContext(ctx context.Context) *Claims {
	c, _ := ctx.Value(ctxKey{}).(*Claims)
	return c
}

func HasRole(ctx context.Context, required ...string) bool {
	roles, ok := UserRolesFromContext(ctx)
	if !ok || len(roles) == 0 || len(required) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		set[r] = struct{}{}
	}
	for _, req := range required {
		if _, ok := set[req]; ok {
			return true
		}
	}
	return false
}

type keycloakRoundTripper struct {
	base   http.RoundTripper
	client *ServiceClient
}

func newKeycloakRoundTripper(base http.RoundTripper, client *ServiceClient) *keycloakRoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &keycloakRoundTripper{base: base, client: client}
}

func (t *keycloakRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.client.GetToken(req.Context())
	if err != nil {
		return nil, fmt.Errorf("service auth: %w", err)
	}
	req2 := req.Clone(req.Context())
	req2.Header.Set("Authorization", "Bearer "+token)
	return t.base.RoundTrip(req2)
}
