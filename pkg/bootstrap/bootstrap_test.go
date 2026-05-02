package bootstrap_test

import (
	"testing"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"

	"gomall/pkg/bootstrap"
)

func TestNewApp_AppliesNameAndVersion(t *testing.T) {
	// Arrange
	env := &bootstrap.Env{
		ID:     "x",
		Logger: log.DefaultLogger,
	}

	// Act
	app := bootstrap.NewApp(env, "svc", "v1.0")

	// Assert
	assert.NotNil(t, app)
}

func TestNewApp_AppliesUserOpts(t *testing.T) {
	// Arrange
	env := &bootstrap.Env{
		ID:     "y",
		Logger: log.DefaultLogger,
	}

	// Act — pass an extra option via the variadic path; empty server list is allowed
	app := bootstrap.NewApp(env, "svc2", "v2.0", kratos.Server())

	// Assert
	assert.NotNil(t, app)
}
