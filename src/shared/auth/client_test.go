package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceClient_GetToken_Caches(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "tok_" + time.Now().Format("150405.000"),
			"expires_in":   300,
		})
	}))
	defer srv.Close()

	client := NewServiceClient(ServiceClientConfig{
		KeycloakConfig: KeycloakConfig{
			RealmURL:     srv.URL + "/realms/go-mall",
			ClientID:     "cart",
			ClientSecret: "secret",
		},
	})

	tok1, err := client.GetToken(t.Context())
	require.NoError(t, err)
	tok2, err := client.GetToken(t.Context())
	require.NoError(t, err)
	assert.Equal(t, tok1, tok2)
	assert.Equal(t, 1, calls)
}

func TestServiceClient_GetToken_RefreshesWhenExpired(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "tok_" + string(rune('0'+callCount)),
			"expires_in":   1,
		})
	}))
	defer srv.Close()

	client := NewServiceClient(ServiceClientConfig{
		KeycloakConfig: KeycloakConfig{
			RealmURL:     srv.URL + "/realms/go-mall",
			ClientID:     "cart",
			ClientSecret: "secret",
		},
	})

	tok1, err := client.GetToken(t.Context())
	require.NoError(t, err)

	client.cacheMu.Lock()
	client.cache.expiresAt = time.Now().Add(-1 * time.Second)
	client.cacheMu.Unlock()

	tok2, err := client.GetToken(t.Context())
	require.NoError(t, err)
	assert.NotEqual(t, tok1, tok2)
}

func TestServiceClient_GetToken_Concurrent(t *testing.T) {
	calls := 0
	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		mu.Unlock()
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "tok_concurrent",
			"expires_in":   300,
		})
	}))
	defer srv.Close()

	client := NewServiceClient(ServiceClientConfig{
		KeycloakConfig: KeycloakConfig{
			RealmURL:     srv.URL + "/realms/go-mall",
			ClientID:     "cart",
			ClientSecret: "secret",
		},
	})

	var wg sync.WaitGroup
	results := make([]string, 10)
	for i := range 10 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			tok, err := client.GetToken(t.Context())
			require.NoError(t, err)
			results[idx] = tok
		}(i)
	}
	wg.Wait()

	mu.Lock()
	assert.Equal(t, 1, calls)
	mu.Unlock()

	for _, r := range results {
		assert.Equal(t, "tok_concurrent", r)
	}
}

func TestServiceClient_GetToken_KeycloakError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := NewServiceClient(ServiceClientConfig{
		KeycloakConfig: KeycloakConfig{
			RealmURL:     srv.URL + "/realms/go-mall",
			ClientID:     "cart",
			ClientSecret: "bad",
		},
	})

	_, err := client.GetToken(t.Context())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}
