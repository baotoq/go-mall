package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ServiceClientConfig struct {
	KeycloakConfig KeycloakConfig
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type cachedToken struct {
	token     string
	expiresAt time.Time
}

type ServiceClient struct {
	config  ServiceClientConfig
	cache   cachedToken
	cacheMu sync.RWMutex
	client  *http.Client
}

func NewServiceClient(config ServiceClientConfig) *ServiceClient {
	return &ServiceClient{
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *ServiceClient) GetToken(ctx context.Context) (string, error) {
	s.cacheMu.RLock()
	if s.cache.token != "" && time.Now().Add(30*time.Second).Before(s.cache.expiresAt) {
		token := s.cache.token
		s.cacheMu.RUnlock()
		return token, nil
	}
	s.cacheMu.RUnlock()

	return s.refreshToken(ctx)
}

func (s *ServiceClient) refreshToken(ctx context.Context) (string, error) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	if s.cache.token != "" && time.Now().Add(30*time.Second).Before(s.cache.expiresAt) {
		return s.cache.token, nil
	}

	data := strings.NewReader(fmt.Sprintf(
		"grant_type=client_credentials&client_id=%s&client_secret=%s",
		s.config.KeycloakConfig.ClientID,
		s.config.KeycloakConfig.ClientSecret,
	))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.tokenEndpoint(), data)
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("keycloak unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status %d", resp.StatusCode)
	}

	var tok tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}

	s.cache = cachedToken{
		token:     tok.AccessToken,
		expiresAt: time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second),
	}

	return tok.AccessToken, nil
}

func (s *ServiceClient) tokenEndpoint() string {
	return fmt.Sprintf("%s/protocol/openid-connect/token", s.config.KeycloakConfig.RealmURL)
}
