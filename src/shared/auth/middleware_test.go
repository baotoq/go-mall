package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockValidator struct {
	claims *Claims
	err    error
}

func (m *mockValidator) Validate(ctx context.Context, token string) (*Claims, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.claims, nil
}

func TestWithClaims(t *testing.T) {
	c := &Claims{Subject: "u1", Roles: []string{"user"}, Email: "a@b.com", Name: "A"}
	ctx := WithClaims(context.Background(), c)
	got := ClaimsFromContext(ctx)
	require.NotNil(t, got)
	assert.Equal(t, "u1", got.Subject)
}

func TestClaimsFromContext_Missing(t *testing.T) {
	assert.Nil(t, ClaimsFromContext(context.Background()))
}

func TestUserIDFromContext(t *testing.T) {
	ctx := WithClaims(context.Background(), &Claims{Subject: "u2"})
	id, ok := UserIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "u2", id)

	id, ok = UserIDFromContext(context.Background())
	assert.False(t, ok)
	assert.Empty(t, id)
}

func TestUserRolesFromContext(t *testing.T) {
	ctx := WithClaims(context.Background(), &Claims{Roles: []string{"admin", "user"}})
	roles, ok := UserRolesFromContext(ctx)
	assert.True(t, ok)
	assert.Len(t, roles, 2)

	_, ok = UserRolesFromContext(context.Background())
	assert.False(t, ok)
}

func TestHasRole(t *testing.T) {
	ctx := WithClaims(context.Background(), &Claims{Roles: []string{"user", "editor"}})
	assert.True(t, HasRole(ctx, "user"))
	assert.True(t, HasRole(ctx, "editor"))
	assert.False(t, HasRole(ctx, "admin"))
	assert.False(t, HasRole(context.Background(), "user"))
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    string
		wantErr string
	}{
		{"valid", "Bearer tok123", "tok123", ""},
		{"missing header", "", "", "missing authorization header"},
		{"wrong scheme", "Basic dXNlcjpwYXNz", "", "invalid authorization header format"},
		{"missing token", "Bearer   ", "", "missing bearer token"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			tok, err := extractBearerToken(req)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, tok)
		})
	}
}

func TestHasAnyRole(t *testing.T) {
	assert.True(t, hasAnyRole([]string{"a", "b"}, []string{"b"}))
	assert.False(t, hasAnyRole([]string{"a"}, []string{"b"}))
}

func TestRequireAuth(t *testing.T) {
	validator := &mockValidator{claims: &Claims{Subject: "u1", Roles: []string{"user"}}}
	middleware := RequireAuth(validator)

	t.Run("valid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer valid")
		rr := httptest.NewRecorder()
		middleware(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("missing header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		middleware(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		})(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		badValidator := &mockValidator{err: errors.New("bad token")}
		mw := RequireAuth(badValidator)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer bad")
		rr := httptest.NewRecorder()
		mw(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		})(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestRequireRole(t *testing.T) {
	validator := &mockValidator{claims: &Claims{Subject: "u1", Roles: []string{"user"}}}
	middleware := RequireRole(validator, "admin")

	t.Run("missing role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer tok")
		rr := httptest.NewRecorder()
		middleware(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		})(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("has role", func(t *testing.T) {
		v := &mockValidator{claims: &Claims{Subject: "u1", Roles: []string{"admin"}}}
		mw := RequireRole(v, "admin")
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer tok")
		rr := httptest.NewRecorder()
		mw(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestRequireServiceAuth(t *testing.T) {
	validator := &mockValidator{claims: &Claims{Subject: "service"}}
	middleware := RequireServiceAuth(validator)

	t.Run("valid service token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments", nil)
		req.Header.Set("Authorization", "Bearer svc")
		rr := httptest.NewRecorder()
		middleware(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
