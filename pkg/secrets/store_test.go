package secrets

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
)

type mockSecretStore struct {
	calls       int
	getSecretFn func(storeName, key string) (map[string]string, error)
}

func (m *mockSecretStore) GetSecret(_ context.Context, storeName, key string, _ map[string]string) (map[string]string, error) {
	m.calls++
	return m.getSecretFn(storeName, key)
}

func noopLogger() log.Logger { return log.NewStdLogger(io.Discard) }

func TestLoadSecrets_Success(t *testing.T) {
	// Arrange
	want := map[string]string{"DATABASE_CONNECTION_STRING": "postgres://localhost/db"}
	mc := &mockSecretStore{getSecretFn: func(_, _ string) (map[string]string, error) {
		return want, nil
	}}

	// Act
	got, err := LoadSecrets(mc, noopLogger(), WithMaxAttempts(1), WithInterval(0))

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, 1, mc.calls)
}

func TestLoadSecrets_RetryThenSuccess(t *testing.T) {
	// Arrange
	want := map[string]string{"key": "val"}
	attempt := 0
	mc := &mockSecretStore{getSecretFn: func(_, _ string) (map[string]string, error) {
		attempt++
		if attempt == 1 {
			return nil, errors.New("not ready")
		}
		return want, nil
	}}

	// Act
	got, err := LoadSecrets(mc, noopLogger(), WithMaxAttempts(2), WithInterval(0))

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, 2, mc.calls)
}

func TestLoadSecrets_Exhausted(t *testing.T) {
	// Arrange
	mc := &mockSecretStore{getSecretFn: func(_, _ string) (map[string]string, error) {
		return nil, errors.New("not ready")
	}}

	// Act
	_, err := LoadSecrets(mc, noopLogger(), WithMaxAttempts(1), WithInterval(0))

	// Assert
	assert.ErrorContains(t, err, "secret store not ready after 5s")
	assert.Equal(t, 1, mc.calls)
}

func TestLoadSecrets_CustomStoreAndKey(t *testing.T) {
	// Arrange
	var capturedStore, capturedKey string
	mc := &mockSecretStore{getSecretFn: func(storeName, key string) (map[string]string, error) {
		capturedStore = storeName
		capturedKey = key
		return map[string]string{}, nil
	}}

	// Act
	_, err := LoadSecrets(mc, noopLogger(), WithStoreName("vault"), WithKey("app-secrets"), WithInterval(0))

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "vault", capturedStore)
	assert.Equal(t, "app-secrets", capturedKey)
}
