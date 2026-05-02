package dapr

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	daprcommon "github.com/dapr/go-sdk/service/common"
	"github.com/stretchr/testify/assert"
)

func TestLoopbackOnly(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cases := []struct {
		name       string
		remoteAddr string
		wantStatus int
	}{
		{"ipv4 loopback", "127.0.0.1:1234", http.StatusOK},
		{"ipv6 loopback", "[::1]:5678", http.StatusOK},
		{"ipv4-mapped ipv6 loopback", "[::ffff:127.0.0.1]:9000", http.StatusOK},
		{"cluster ip", "10.0.0.1:1234", http.StatusForbidden},
		{"external ip", "203.0.113.1:80", http.StatusForbidden},
		{"empty remote addr", "", http.StatusForbidden},
		{"malformed remote addr", "not-an-addr", http.StatusForbidden},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			req := httptest.NewRequest(http.MethodPost, "/dapr/events/test", nil)
			req.RemoteAddr = tc.remoteAddr
			w := httptest.NewRecorder()

			// Act
			LoopbackOnly(ok)(w, req)

			// Assert
			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestTopicHandler_panicInHandler_returns500(t *testing.T) {
	// Arrange — a handler that panics
	panicking := daprcommon.TopicEventHandler(func(_ context.Context, _ *daprcommon.TopicEvent) (bool, error) {
		panic("simulated handler panic")
	})
	body := `{"id":"e1","source":"test","type":"test","specversion":"1.0","datacontenttype":"application/json","data":{}}`
	req := httptest.NewRequest(http.MethodPost, "/dapr/events/test", strings.NewReader(body))
	w := httptest.NewRecorder()

	// Act — must not propagate panic
	assert.NotPanics(t, func() {
		TopicHandler(panicking)(w, req)
	})

	// Assert — Dapr interprets 500 as RETRY
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
