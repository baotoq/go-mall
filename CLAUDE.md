# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
make build                          # outputs to ./bin/tradingbot

# Run
./bin/tradingbot -conf ./app/tradingbot/configs
go run ./app/tradingbot/cmd/server -conf ./app/tradingbot/configs

# Generate protobuf (API + HTTP + gRPC stubs)
make api                            # from api/**/*.proto
make config                         # from app/**/*.proto (conf structs)

# Re-generate wire DI
cd app/tradingbot/cmd/server && wire

# Generate all
make all                            # api + config + go generate + go mod tidy

# Install code-gen tools
make init
```

Tests: `go test ./...` (no custom test runner).

```bash
# Dev environment (requires Tilt + k8s context docker-desktop or orbstack)
make dev                            # tilt up — hot-reload via binary sync, no image rebuild
make debug                          # tilt up --debug — attaches Delve on :2345
```

## Architecture

go-kratos v2 layered monolith with Wire DI. Request flow:

```
proto (api/) → service → biz (usecase) → data (repo)
```

| Layer | Package | Responsibility |
|-------|---------|---------------|
| **server** | `app/tradingbot/internal/server` | HTTP (`:8000`) + gRPC (`:9000`) server setup, middleware |
| **service** | `app/tradingbot/internal/service` | Implements proto-generated server interfaces, delegates to biz |
| **biz** | `app/tradingbot/internal/biz` | Domain logic, defines repo interfaces (`GreeterRepo`), owns domain errors |
| **data** | `app/tradingbot/internal/data` | Implements biz repo interfaces, holds DB/Redis clients |
| **conf** | `app/tradingbot/internal/conf` | Config structs generated from `conf.proto`; loaded via `app/tradingbot/configs/config.yaml` |
| **api** | `api/` | Proto definitions; generated `.pb.go`, `_grpc.pb.go`, `_http.pb.go` |

### Key conventions

- **Dependency direction**: `server → service → biz ← data` (data depends on biz interfaces, not the other way).
- **Wire**: each layer exposes a `ProviderSet`. Add new constructors there when adding types. Regenerate with `wire` after changes to `wire.go`.
- **Errors**: define in `biz` using `errors.NotFound`/`errors.BadRequest` etc. from `go-kratos/kratos/v2/errors`. Map proto `ErrorReason` enums for typed errors.
- **Config**: add fields to `app/tradingbot/internal/conf/conf.proto`, run `make config` to regenerate, then read via `*conf.Data` or `*conf.Server` injected by Wire.
- **New API endpoints**: define in `api/**/*.proto`, run `make api`, implement in `app/tradingbot/internal/service`.

### ORM (ent)

`ent/` is generated from schemas in `app/tradingbot/internal/data/ent/schema/`. After changing a schema, regenerate with `go generate ./app/tradingbot/internal/data/ent/...` (or `make generate`). The client is wired through `*data.Data`. Add new entities as schemas there.

### Infrastructure

Dev stack (via Tilt + Helm): **Postgres** (`:5432`) + **Redis** (`:6379`) deployed to local k8s. App ports: HTTP `:8000`, gRPC `:9000`, Delve `:2345` (debug only). Overlays in `deploy/k8s/overlays/local` and `deploy/k8s/overlays/debug`.
