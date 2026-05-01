package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	w := httptest.NewRecorder()
	Healthz(w, &http.Request{})

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if body := w.Body.String(); body != "I'm alive" {
		t.Errorf("expected body %q, got %q", "I'm alive", body)
	}
}
