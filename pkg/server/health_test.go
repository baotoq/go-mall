package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthz(t *testing.T) {
	// Arrange
	w := httptest.NewRecorder()
	req := &http.Request{}

	// Act
	Healthz(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "I'm alive", w.Body.String())
}
