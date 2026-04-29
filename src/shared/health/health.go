package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/rest"
)

// Probe checks whether a dependency is healthy.
type Probe interface {
	Name() string
	Check(ctx context.Context) error
}

// Register adds /healthz (liveness) and /readyz (readiness) routes to the server.
func Register(server *rest.Server, probes ...Probe) {
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/healthz", Handler: healthzHandler()},
		{Method: http.MethodGet, Path: "/readyz", Handler: readyzHandler(probes)},
	})
}

func healthzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

type failedProbe struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

func readyzHandler(probes []Probe) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		var failed []failedProbe
		for _, p := range probes {
			if err := p.Check(ctx); err != nil {
				failed = append(failed, failedProbe{Name: p.Name(), Error: err.Error()})
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if len(failed) > 0 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "unready", "failed": failed})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
