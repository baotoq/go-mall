// Package bootstrap handles the shared startup sequence for every service:
// logger construction, config loading, Dapr sidecar connection, and secret
// fetching. Steps specific to each service (Scan, wireApp, Run) remain in
// main.go.
//
// Flag parsing is intentionally NOT handled here. Each service's main()
// registers and parses its own flags before calling Init, then passes the
// resolved config path via Options.ConfigPath. This keeps bootstrap free of
// global flag state and safe for concurrent test invocation.
package bootstrap

import (
	"context"
	"errors"
	"os"
	"sync"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"

	"gomall/pkg/secrets"
)

// Options configures Init.
type Options struct {
	Name       string // service name (required, used in logger fields)
	Version    string // service version (optional, defaults to "")
	ConfigPath string // required; path passed to kratos config/file source
}

// Env is what main() needs after bootstrap is done.
//
// Lifetime: Config and Dapr are owned by the cleanup func returned from Init.
// Once cleanup() has been called, neither field may be used. Callers that
// need post-cleanup access must copy out values (e.g. Scan into a local
// struct) before invoking cleanup.
type Env struct {
	ID      string            // os.Hostname() or "unknown"
	Logger  log.Logger        // kratos logger with standard fields
	Config  config.Config     // already loaded; caller does c.Scan(&bc)
	Dapr    dapr.Client       // connected sidecar client
	Secrets map[string]string // raw secret map; caller does secrets.Parse
}

// Init builds the logger, loads config, dials Dapr, and fetches secrets.
// Returns Env plus a cleanup func (closes dapr then config, LIFO). Any error
// is returned to the caller; Init never panics.
//
// ctx is forwarded to secret-fetch retries so the caller can bound the total
// startup time (LoadSecrets retries up to ~60s by default).
func Init(ctx context.Context, opts Options) (*Env, func(), error) {
	if opts.Name == "" {
		return nil, nil, errors.New("bootstrap: Options.Name is required")
	}
	if opts.ConfigPath == "" {
		return nil, nil, errors.New("bootstrap: Options.ConfigPath is required")
	}

	// Resolve service ID with explicit "unknown" fallback.
	id, err := os.Hostname()
	if err != nil || id == "" {
		id = "unknown"
	}

	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", opts.Name,
		"service.version", opts.Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	c := config.New(
		config.WithSource(
			file.NewSource(opts.ConfigPath),
		),
	)
	if err := c.Load(); err != nil {
		c.Close() //nolint:errcheck
		return nil, nil, err
	}

	daprClient, err := dapr.NewClient()
	if err != nil {
		c.Close() //nolint:errcheck
		return nil, nil, err
	}

	if err := ctx.Err(); err != nil {
		daprClient.Close()
		c.Close() //nolint:errcheck
		return nil, nil, err
	}

	secret, err := secrets.LoadSecrets(daprClient, logger)
	if err != nil {
		daprClient.Close()
		c.Close() //nolint:errcheck
		return nil, nil, err
	}

	env := &Env{
		ID:      id,
		Logger:  logger,
		Config:  c,
		Dapr:    daprClient,
		Secrets: secret,
	}

	var cleanOnce sync.Once
	cleanup := func() {
		cleanOnce.Do(func() {
			daprClient.Close()
			c.Close() //nolint:errcheck
		})
	}

	return env, cleanup, nil
}

// NewApp builds a kratos.App with the standard ID/Name/Version/Logger/Metadata
// wiring, then applies caller-supplied options (Server, BeforeStart, AfterStop,
// etc.). Use this from main() after wire returns the kratos servers.
func NewApp(env *Env, name, version string, opts ...kratos.Option) *kratos.App {
	base := []kratos.Option{
		kratos.ID(env.ID),
		kratos.Name(name),
		kratos.Version(version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(env.Logger),
	}
	return kratos.New(append(base, opts...)...)
}
