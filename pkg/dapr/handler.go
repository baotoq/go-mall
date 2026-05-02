package dapr

import (
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"

	daprcommon "github.com/dapr/go-sdk/service/common"
)

// Subscription is one entry in the /dapr/subscribe JSON response.
type Subscription struct {
	PubsubName string `json:"pubsubname"`
	Topic      string `json:"topic"`
	Route      string `json:"route"`
}

// cloudEvent is the CloudEvents 1.0 envelope the Dapr sidecar sends on delivery.
type cloudEvent struct {
	ID              string          `json:"id"`
	Source          string          `json:"source"`
	Type            string          `json:"type"`
	SpecVersion     string          `json:"specversion"`
	DataContentType string          `json:"datacontenttype"`
	Data            json.RawMessage `json:"data"`
	DataBase64      string          `json:"data_base64,omitempty"`
	Topic           string          `json:"topic"`
	PubsubName      string          `json:"pubsubname"`
	Subject         string          `json:"subject"`
	TraceID         string          `json:"traceid"`
	TraceParent     string          `json:"traceparent"`
}

// maxCloudEventBytes caps the request body to prevent unbounded memory allocation.
const maxCloudEventBytes = 1 << 20 // 1 MiB

// TopicHandler wraps a Dapr TopicEventHandler as a net/http.HandlerFunc.
// It decodes the incoming CloudEvent envelope, populates a *common.TopicEvent
// (including RawData for inbox deduplication and TypedHandler unmarshalling),
// invokes the handler, and writes the Dapr ACK/RETRY/DROP response.
//
// Routes registered via TopicHandler bypass Kratos middleware (they are raw
// HandleFunc routes, not proto-generated handlers). This is intentional: the
// Dapr sidecar carries no user JWT, so these routes must be unauthenticated
// at the app layer. Restrict access at the network/pod level instead.
func TopicHandler(h daprcommon.TopicEventHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxCloudEventBytes)
		var ce cloudEvent
		if err := json.NewDecoder(r.Body).Decode(&ce); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		evt := &daprcommon.TopicEvent{
			ID:              ce.ID,
			Source:          ce.Source,
			Type:            ce.Type,
			SpecVersion:     ce.SpecVersion,
			DataContentType: ce.DataContentType,
			Topic:           ce.Topic,
			PubsubName:      ce.PubsubName,
			Subject:         ce.Subject,
			TraceID:         ce.TraceID,
			TraceParent:     ce.TraceParent,
			RawData:         ce.Data,
		}
		if len(evt.RawData) == 0 && len(ce.DataBase64) > 0 {
			decoded, err := base64.StdEncoding.DecodeString(ce.DataBase64)
			if err != nil {
				http.Error(w, "invalid data_base64", http.StatusBadRequest)
				return
			}
			evt.RawData = decoded
		}

		retry, err := h(r.Context(), evt)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			status := "DROP"
			if retry {
				status = "RETRY"
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"status": status})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "SUCCESS"})
	}
}

// SubscribeHandler returns an http.HandlerFunc that serves GET /dapr/subscribe.
func SubscribeHandler(subs []Subscription) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(subs)
	}
}

// LoopbackOnly rejects requests not originating from the loopback interface.
// The Dapr sidecar always delivers via the pod-local loopback (127.0.0.1/::1),
// so this guard ensures no external caller can reach Dapr delivery routes.
func LoopbackOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if ip := net.ParseIP(host); ip == nil || !ip.IsLoopback() {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}
