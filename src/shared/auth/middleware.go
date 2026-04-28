package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/core/logx"
)

type contextKey string

const (
	userIDKey    contextKey = "user_id"
	userRolesKey contextKey = "user_roles"
)

type Claims struct {
	Subject string
	Roles   []string
	Email   string
	Name    string
}

type TokenValidator interface {
	Validate(ctx context.Context, token string) (*Claims, error)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	if c := ClaimsFromContext(ctx); c != nil {
		return c.Subject, true
	}
	val := ctx.Value(userIDKey)
	if val == nil {
		return "", false
	}
	id, ok := val.(string)
	return id, ok
}

func UserRolesFromContext(ctx context.Context) ([]string, bool) {
	if c := ClaimsFromContext(ctx); c != nil {
		return c.Roles, true
	}
	val := ctx.Value(userRolesKey)
	if val == nil {
		return nil, false
	}
	roles, ok := val.([]string)
	return roles, ok
}

func AuthMiddleware(validator TokenValidator) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			token, err := extractBearerToken(r)
			if err != nil {
				writeError(ctx, w, http.StatusUnauthorized, err.Error())
				return
			}

			claims, err := validator.Validate(ctx, token)
			if err != nil {
				writeError(ctx, w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx = context.WithValue(ctx, userIDKey, claims.Subject)
			ctx = context.WithValue(ctx, userRolesKey, claims.Roles)

			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

func RequireAuth(validator TokenValidator) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			token, err := extractBearerToken(r)
			if err != nil {
				writeError(ctx, w, http.StatusUnauthorized, err.Error())
				return
			}

			claims, err := validator.Validate(ctx, token)
			if err != nil {
				writeError(ctx, w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx = context.WithValue(ctx, userIDKey, claims.Subject)
			ctx = context.WithValue(ctx, userRolesKey, claims.Roles)

			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

func RequireRole(validator TokenValidator, roles ...string) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			token, err := extractBearerToken(r)
			if err != nil {
				writeError(ctx, w, http.StatusUnauthorized, err.Error())
				return
			}

			claims, err := validator.Validate(ctx, token)
			if err != nil {
				writeError(ctx, w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			if !hasAnyRole(claims.Roles, roles) {
				writeError(ctx, w, http.StatusForbidden, "insufficient permissions")
				return
			}

			ctx = context.WithValue(ctx, userIDKey, claims.Subject)
			ctx = context.WithValue(ctx, userRolesKey, claims.Roles)

			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

func RequireServiceAuth(validator TokenValidator) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			token, err := extractBearerToken(r)
			if err != nil {
				writeError(ctx, w, http.StatusUnauthorized, err.Error())
				return
			}

			claims, err := validator.Validate(ctx, token)
			if err != nil {
				logx.Errorf("service auth failed: %v", err)
				writeError(ctx, w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx = context.WithValue(ctx, userIDKey, claims.Subject)
			ctx = context.WithValue(ctx, userRolesKey, claims.Roles)

			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

func extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid authorization header format")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("missing bearer token")
	}

	return token, nil
}

func hasAnyRole(userRoles, requiredRoles []string) bool {
	for _, userRole := range userRoles {
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				return true
			}
		}
	}
	return false
}

func writeError(ctx context.Context, w http.ResponseWriter, code int, message string) {
	logx.WithContext(ctx).Error(message)
	httpx.WriteJsonCtx(ctx, w, code, map[string]any{
		"code":    code,
		"message": message,
	})
}