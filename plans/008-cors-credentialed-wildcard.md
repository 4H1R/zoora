# Plan 008: Harden default CORS — never pair `*` with credentials

> **Executor instructions**: Follow step by step. Verify each step. Honor STOP
> conditions. Update this plan's row in `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat 0071d2e..HEAD -- internal/middleware/cors.go internal/config/config.go cmd/api/main.go`
> Mismatch vs "Current state" = STOP.

## Status

- **Priority**: P2
- **Effort**: S
- **Risk**: LOW
- **Depends on**: none
- **Category**: security
- **Planned at**: commit `0071d2e`, 2026-07-21

## Why this matters

`CORS_ALLOWED_ORIGINS` defaults to `*` (`internal/config/config.go:58`), and the CORS
middleware sets `AllowCredentials: true` (`internal/middleware/cors.go:20`). A credentialed
wildcard is a browser-trust misconfiguration: any deployment that forgets to set an
explicit origin list runs with "allow any origin, with credentials". The API currently
authenticates via `Authorization: Bearer` (not cookies), which limits today's blast radius,
but relying on ops to always override an insecure default is fragile and any future
cookie-bearing surface would be immediately exposed. The fix makes the insecure combination
impossible: credentials are only enabled for an explicit allow-list, never for `*`.

## Current state

`internal/config/config.go:58`:
```go
CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"*"`
```
`internal/middleware/cors.go`:
```go
func CORS(allowedOrigins []string) gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins:     allowedOrigins,
        AllowWildcard:    true,        // needed for "https://*.zoora.ir" tenant hosts
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    })
}
```
Wired at `cmd/api/main.go:275`: `middleware.CORS(cfg.CORSAllowedOrigins)`.

Exemplar for a production-only safety override — `cmd/api/main.go:256`:
```go
rateLimitDisabled := cfg.RateLimitDisabled && !cfg.IsProduction()
```
(An unsafe escape-hatch is neutralized in production. Mirror this intent for CORS.)

Note: `https://*.zoora.ir` (a specific subdomain wildcard) is **by design** and safe with
credentials — gin-contrib reflects the concrete matched origin. The problem is only the
bare `*`.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Build | `go build ./...` | exit 0 |
| Tests (middleware) | `go test -race -count=1 ./internal/middleware/...` | all pass |
| Tests (config) | `go test -race -count=1 ./internal/config/...` | all pass (see STOP note) |
| Lint | `make lint` | exit 0 |

## Scope

**In scope**:
- `internal/middleware/cors.go`
- `internal/middleware/*_test.go` (add a CORS test)
- `.env.example` and `.env.prod.example` (documentation only — clarify the var; do NOT touch real `.env`/`.env.prod`)

**Out of scope**:
- Real `.env` / `.env.prod` (gitignored secrets) — never modify.
- The tenant `https://*.zoora.ir` wildcard behavior — must keep working with credentials.
- `internal/config/config.go` default — leave the `*` default (dev convenience); the middleware neutralizes its danger. (Changing the default is an alternative, but riskier for local dev; keep it out unless STOP-noted.)

## Git workflow

- Branch: `advisor/008-cors-credentialed-wildcard`
- Conventional commits. No push/PR unless instructed.

## Steps

### Step 1: Disable credentials whenever origins include a bare `*`

In `internal/middleware/cors.go`, compute `allowCredentials` = true unless the list
contains a bare `"*"`:
```go
func CORS(allowedOrigins []string) gin.HandlerFunc {
    allowCredentials := true
    for _, o := range allowedOrigins {
        if o == "*" {
            allowCredentials = false // a credentialed wildcard is insecure (and browsers reject it anyway)
            break
        }
    }
    return cors.New(cors.Config{
        AllowOrigins:     allowedOrigins,
        AllowWildcard:    true,
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
        AllowCredentials: allowCredentials,
        MaxAge:           12 * time.Hour,
    })
}
```
An explicit allow-list (including the `https://*.zoora.ir` subdomain wildcard) keeps
`AllowCredentials: true`; only a bare `*` drops it.

**Verify**: `go build ./...` → exit 0.

### Step 2: Document the variable

In `.env.example` and `.env.prod.example`, ensure `CORS_ALLOWED_ORIGINS` has a comment:
set an explicit comma-separated origin list in production (e.g. `https://*.zoora.ir`); a
bare `*` disables credentialed CORS.

**Verify**: `git diff --name-only` shows only example files (not real env files).

### Step 3: Run suites

**Verify**:
- `go test -race -count=1 ./internal/middleware/...` → all pass
- `make lint` → exit 0

## Test plan

Add a middleware test (model after existing `internal/middleware/*_test.go`, using
`httptest` + a gin engine with `CORS(...)` mounted). Cases:
- Origins `["*"]`: a preflight/response does **not** include `Access-Control-Allow-Credentials: true`.
- Origins `["https://*.zoora.ir"]`: a request from `https://acme.zoora.ir` is allowed **and** carries `Access-Control-Allow-Credentials: true`.

Verification: `go test -race -count=1 ./internal/middleware/...` → all pass, new cases included.

## Done criteria

- [ ] `go build ./...` exits 0
- [ ] `go test -race -count=1 ./internal/middleware/...` exits 0 with new CORS tests passing
- [ ] `make lint` exits 0
- [ ] With `CORS_ALLOWED_ORIGINS=*`, credentials are NOT allowed; with an explicit list, they are
- [ ] Real `.env` / `.env.prod` untouched (`git status`)
- [ ] `plans/README.md` row for 008 updated

## STOP conditions

- Excerpts don't match live code (drift).
- gin-contrib/cors `cors.New` panics at startup for `["*"]` with `AllowWildcard: true` in this version — then a bare `*` was never actually functional and the config default should change to an explicit dev list instead; STOP and report the panic.
- `go test ./internal/config/...` was already failing before your change (a known env-dependent test) — do not attribute it to this plan; note it and proceed. (Per repo memory, one config test fails inside the container but passes on host.)
- Verification fails twice after a reasonable fix.

## Maintenance notes

- Reviewer: confirm the tenant `https://*.zoora.ir` path still returns credentials (multi-tenant cookies/future sessions depend on it).
- Better end-state (deferred): fail-closed in production if `CORS_ALLOWED_ORIGINS` is unset or `*`, mirroring the `rateLimitDisabled` production override at `main.go:256`. Left out here to avoid changing startup behavior; revisit if a cookie-based surface is added.
