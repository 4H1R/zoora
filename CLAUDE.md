# Project

Zoora — a video conferencing, virtual classroom, and LMS (learning management system) SaaS (Multi Tenant). Monorepo: Go backend at the repo root plus a `frontend/` folder with the web client.

The backend builds two binaries from the same codebase: API server (`cmd/api`) and Asynq worker (`cmd/worker`).

This file covers the **backend**. For frontend work (stack, commands, conventions) see `frontend/CLAUDE.md`.

## Commands

```bash
make run-api              # Run API server
make run-worker           # Run Asynq background worker
make test                 # Unit tests: go test ./internal/... -v -race -count=1
make test-integration     # Integration tests (testcontainers): go test -tags=integration ./tests/... -v -race -count=1
make lint                 # golangci-lint run ./...
make docker-up            # Start local deps (Postgres, Redis, RustFS/S3, LiveKit)
make docker-down          # Stop local deps
make migrate-up           # Run all pending migrations
make migrate-reset        # Drop DB and re-run all migrations (fresh start)
make migrate-create name=<name>  # Create new up-only migration
```

Run a single test: `go test -v -race -count=1 -run TestName ./internal/users/...`

Run a single integration test: `go test -tags=integration -v -race -count=1 -run TestName ./tests/integration/...`

## Architecture

**Feature-first layout** with `internal/domain/` as the dependency root.

### Dependency rules
- `domain/` imports nothing from the project — models, DTOs, interfaces, sentinel errors, and task type constants all live here
- Feature packages (`users/`, `meetings/`, `recordings/`) import only `domain/`, `platform/`, and `config/` — never each other. Cross-feature interaction uses domain interfaces injected via constructors
- `platform/` packages import only `config/` and external libs — no business logic

### Layers per feature (Handler → Service → Repository)
- **Handler**: HTTP binding/validation, maps domain errors to HTTP status codes, returns standardized JSON (`{"success": bool, "data"|"error": ...}`)
- **Service**: implements domain interface, orchestrates repo + platform clients (queue, LiveKit, S3), performs authz checks, returns domain errors
- **Repository**: GORM queries only, receives/returns domain models

### Background jobs
- API enqueues tasks via Asynq client; worker processes them via Asynq server
- Task type constants and payload structs live in `domain/queue.go`
- Task handler functions live in feature packages (`meetings/task_handlers.go`, `recordings/task_handlers.go`), registered in `cmd/worker/main.go`

## API codegen workflow

When adding, modifying, or removing routes/handlers (including swagger annotations):
1. `make generate` — regenerate OpenAPI spec via swag

## Code search
**IMPORTANT**

Prefer semble MCP tools (`mcp__semble__search`, `mcp__semble__find_related`) over grep/glob for exploring or locating code. Use grep only when semble is unavailable or for exact literal matches.

## Key conventions

- Config loaded from env vars into typed struct (`internal/config/config.go`) — no global state
- All deps injected via constructors, no `init()` for business logic
- Always propagate `ctx context.Context` as first param through all layers
- Wrap errors with context: `fmt.Errorf("creating meeting: %w", err)`, use `errors.Is()`/`errors.As()`
- Never use GORM AutoMigrate in production — use `golang-migrate` SQL files only
- Integration tests use `testcontainers-go` for real Postgres and Redis
- Local S3-compatible storage is RustFS (not MinIO) — see `docker-compose.yml`
- Makefile sources `.env` — ensure it exists for make targets to work
