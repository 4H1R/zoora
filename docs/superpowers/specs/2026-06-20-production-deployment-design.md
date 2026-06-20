# Production Deployment Design — Zoora

**Date:** 2026-06-20
**Status:** Approved

## Goal

Production-ready Docker images and a `docker-compose.prod.yml` to deploy the Zoora
backend (Go: api + worker) and frontend (Vite/React SPA) on a single VM, fully
self-contained, with automatic TLS.

## Decisions

- **Deps:** self-host all stateful deps (Postgres, Redis, RustFS/S3, LiveKit) as
  containers.
- **Edge/TLS:** Traefik reverse proxy with automatic Let's Encrypt certificates.
- **Frontend:** static SPA served by nginx; calls the API on a separate origin
  (`api.$DOMAIN`) via a build-time-baked `VITE_API_URL`. Requires CORS on the backend.
- **Build model:** images built on the deploy host (`build:` context). No registry, no CI.
- **Domains:** subdomains of one apex (`$DOMAIN`).
- **S3 access:** browser uses direct presigned URLs pointing at `s3.$DOMAIN`.

## Topology

Single VM. `docker-compose.prod.yml`. Traefik terminates TLS on :443 and routes by
host. All routing subdomains derive from one apex `$DOMAIN`.

| Host           | Service              | Exposure                  | Notes                                  |
|----------------|----------------------|---------------------------|----------------------------------------|
| `app.$DOMAIN`  | frontend (nginx)     | Traefik :443              | static SPA, `try_files` fallback       |
| `api.$DOMAIN`  | backend `api`        | Traefik :443              | gin, behind proxy                      |
| `livekit.$DOMAIN` | livekit signaling | Traefik :443 (ws/wss)     | `wss://` from browser                  |
| `s3.$DOMAIN`   | rustfs               | Traefik :443              | presigned URL host                     |
| traefik dashboard | traefik           | disabled by default       | enable behind basic-auth if needed     |

- `worker` exposes no ports (Asynq consumer).
- LiveKit media ports published **directly on host**: `7881/tcp`, `7882/udp`
  (Traefik cannot proxy WebRTC UDP). `node_ip` set to the host's public IP.

## Images (built on host)

### Backend — `Dockerfile` (rewrite existing)
- Multi-stage. Builder = `golang:1.25-alpine` (match `go.mod`).
- Build all three binaries: `api`, `worker`, `seed`
  (`CGO_ENABLED=0`, `-ldflags="-w -s"`).
- Runtime = small base (`alpine:3.x` with `ca-certificates`, `tzdata`). Single image;
  `command:` selects the binary per compose service.
- Copy `migrations/` into image (used by seed/debug; primary migration path is the
  dedicated migrate service below).

### Frontend — `frontend/Dockerfile` (new)
- Stage 1: `node:22-alpine`, `corepack`/`pnpm install --frozen-lockfile`, then
  `pnpm build` with `ARG VITE_API_URL` (and any other `VITE_*` baked at build time).
- Stage 2: `nginx:alpine` serving `/usr/share/nginx/html` from the Vite `dist/`.
- `nginx.conf`: SPA fallback `try_files $uri $uri/ /index.html`, gzip, static asset
  caching headers. No API proxy (separate origin model).

### Migrations — one-shot service
- Image `migrate/migrate:latest`, mounts `./migrations`, runs
  `-path /migrations -database "$DATABASE_URL" up`.
- `api` and `worker` `depends_on:` it with `condition: service_completed_successfully`.

### Bucket init — one-shot service
- `amazon/aws-cli`, points at `http://rustfs:9000`, idempotently creates the
  `$S3_BUCKET` bucket (`s3 mb` ignoring "already exists"). Runs before `api`.
- Needed because the app auto-creates the bucket only in development.

### Seed — one-shot, manual
- Same backend image, `command: seed`, `profiles: ["seed"]`. Run via
  `docker compose --profile seed run --rm seed`. Never runs automatically.

## LiveKit production

- `livekit.yaml` config file (not `--dev`):
  - real `keys:` (`LIVEKIT_API_KEY: LIVEKIT_API_SECRET`),
  - `redis:` pointed at the compose redis,
  - `rtc.node_ip` = host public IP, `rtc.tcp_port: 7881`, `rtc.port_range` / UDP `7882`,
  - `rtc.use_external_ip: true`.
- Traefik routes `livekit.$DOMAIN` → signaling `:7880` (HTTP upgrade to ws).
- Frontend `VITE` / backend `LIVEKIT_PUBLIC_URL = wss://livekit.$DOMAIN`.
- Backend `LIVEKIT_HOST = http://livekit:7880` (internal).

## Backend code changes (minimal)

1. **CORS origins via env.** Add `CORSAllowedOrigins []string` to `internal/config/config.go`
   (`env:"CORS_ALLOWED_ORIGINS"`). Wire into the `middleware.CORS(...)` call in
   `cmd/api/main.go`, replacing the hardcoded `[]string{"*"}`. In production set
   `CORS_ALLOWED_ORIGINS=https://app.$DOMAIN`. Keep `*` as the default for development.
   Rationale: `AllowOrigins:["*"]` with `AllowCredentials:true` is rejected by browsers
   on cross-origin requests.
2. **Trusted proxies.** Configure gin trusted proxies (or `SetTrustedProxies`) so client
   IPs and `X-Forwarded-*` are honored correctly behind Traefik. Restrict to the Docker
   network range.

These are the only application changes; everything else is infrastructure.

## Configuration & secrets

- `.env.prod.example` (committed) documents every production variable:
  - `DOMAIN`, `ACME_EMAIL`
  - `ENVIRONMENT=production`
  - `DATABASE_URL` (internal: `postgres://...@pgsql:5432/...`)
  - `REDIS_URL=redis://redis:6379`
  - `JWT_SECRET` (strong, generated), `JWT_EXPIRY`
  - `LIVEKIT_HOST=http://livekit:7880`, `LIVEKIT_PUBLIC_URL=wss://livekit.$DOMAIN`,
    `LIVEKIT_API_KEY`, `LIVEKIT_API_SECRET`
  - `S3_ENDPOINT=https://s3.$DOMAIN`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`,
    `S3_REGION`
  - `CORS_ALLOWED_ORIGINS=https://app.$DOMAIN`
  - DB/rustfs credentials for the dep containers
  - Frontend build arg `VITE_API_URL=https://api.$DOMAIN`
- Real `.env.prod` is gitignored.
- Persistence via named volumes: `pgsql-data`, `redis-data`, `rustfs-data`, plus a
  Traefik ACME volume (`letsencrypt`).

## Healthchecks & startup ordering

- Postgres (`pg_isready`), Redis (`redis-cli ping`), RustFS (HTTP `/health`) — as in dev.
- `api` healthcheck hits the existing `internal/platform/health` endpoint.
- Startup order via `depends_on` conditions:
  `pgsql/redis healthy` → `migrate` + `bucket-init` complete → `api`/`worker` start.

## Out of scope (YAGNI)

- Kubernetes, multi-node, autoscaling.
- CI/CD pipeline, image registry, image signing.
- Monitoring/observability stack (Prometheus/Grafana/Loki).
- Backups automation (documented manually, not built).
- The existing dev `docker-compose.yml` and `Dockerfile.dev` stay untouched.

## Deliverables

- `Dockerfile` (rewritten, multi-binary, Go 1.25).
- `frontend/Dockerfile` + `frontend/nginx.conf`.
- `docker-compose.prod.yml`.
- `deploy/livekit.yaml` (prod LiveKit config).
- `.env.prod.example`.
- Backend changes: `internal/config/config.go`, `cmd/api/main.go` (CORS env + trusted
  proxies).
- A short `docs/DEPLOYMENT.md` with first-deploy steps (DNS records, `.env.prod`,
  `docker compose -f docker-compose.prod.yml up -d --build`, seed command).
