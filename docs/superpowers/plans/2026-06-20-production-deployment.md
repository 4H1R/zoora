# Production Deployment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship production-ready Docker images and a `docker-compose.prod.yml` to deploy the Zoora backend (api + worker) and frontend (Vite SPA) on a single VM with self-hosted deps and automatic TLS via Traefik.

**Architecture:** One multi-binary backend image (api/worker/seed), one nginx static-SPA frontend image, Traefik edge terminating TLS and routing subdomains of one apex. All stateful deps (Postgres, Redis, RustFS, LiveKit) self-hosted with named volumes. Migrations and bucket creation run as one-shot dependency services. Two small backend code changes (env-driven CORS, trusted proxies).

**Tech Stack:** Go 1.25, Docker multi-stage builds, docker compose, Traefik v3, nginx, node 22 / pnpm, migrate/migrate, LiveKit.

Reference spec: `docs/superpowers/specs/2026-06-20-production-deployment-design.md`

---

## File Structure

- Modify: `internal/config/config.go` — add `CORSAllowedOrigins` field.
- Modify: `cmd/api/main.go` — wire env CORS + set trusted proxies.
- Test: `internal/config/config_test.go` — assert CORS env parsing.
- Rewrite: `Dockerfile` — multi-stage, Go 1.25, builds api/worker/seed.
- Create: `frontend/Dockerfile` — node build → nginx runtime.
- Create: `frontend/nginx.conf` — SPA fallback + gzip + asset caching.
- Create: `deploy/livekit.yaml` — production LiveKit config.
- Create: `docker-compose.prod.yml` — full production stack.
- Create: `.env.prod.example` — documented production env.
- Modify: `.gitignore` — ignore `.env.prod`.
- Create: `docs/DEPLOYMENT.md` — first-deploy runbook.

> Note: `docs/` is gitignored (swagger output). Commit the two docs files with `git add -f`.

---

## Task 1: Backend config — env-driven CORS origins

**Files:**
- Modify: `internal/config/config.go`
- Test: `internal/config/config_test.go` (create if absent)

- [ ] **Step 1: Write the failing test**

Create/append `internal/config/config_test.go`:

```go
package config_test

import (
	"testing"

	"github.com/4H1R/zoora/internal/config"
)

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://x:y@localhost:5432/z?sslmode=disable")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("LIVEKIT_HOST", "http://localhost:7880")
	t.Setenv("LIVEKIT_API_KEY", "devkey")
	t.Setenv("LIVEKIT_API_SECRET", "devsecret")
	t.Setenv("S3_ENDPOINT", "http://localhost:9000")
	t.Setenv("S3_BUCKET", "private")
	t.Setenv("S3_ACCESS_KEY", "dev")
	t.Setenv("S3_SECRET_KEY", "dev")
	t.Setenv("JWT_SECRET", "secret")
}

func TestLoad_CORSAllowedOrigins_FromEnv(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com,https://admin.example.com")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	got := cfg.CORSAllowedOrigins
	want := []string{"https://app.example.com", "https://admin.example.com"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("origin[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestLoad_CORSAllowedOrigins_DefaultWildcard(t *testing.T) {
	setRequiredEnv(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.CORSAllowedOrigins) != 1 || cfg.CORSAllowedOrigins[0] != "*" {
		t.Fatalf("default CORS = %v, want [*]", cfg.CORSAllowedOrigins)
	}
}
```

- [ ] **Step 2: Run test, verify it fails**

Run: `go test ./internal/config/... -run TestLoad_CORSAllowedOrigins -count=1`
Expected: FAIL — compile error `cfg.CORSAllowedOrigins undefined`.

- [ ] **Step 3: Add the field**

In `internal/config/config.go`, add inside the `Config` struct (next to the other S3/JWT fields):

```go
	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"*"`
```

- [ ] **Step 4: Run test, verify it passes**

Run: `go test ./internal/config/... -run TestLoad_CORSAllowedOrigins -count=1`
Expected: PASS (both subtests).

- [ ] **Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): add CORS_ALLOWED_ORIGINS env"
```

---

## Task 2: Wire env CORS + trusted proxies in api

**Files:**
- Modify: `cmd/api/main.go` (around lines 166-176)

- [ ] **Step 1: Replace hardcoded CORS origins**

In `cmd/api/main.go`, change line 171 from:

```go
		middleware.CORS([]string{"*"}),
```

to:

```go
		middleware.CORS(cfg.CORSAllowedOrigins),
```

- [ ] **Step 2: Set trusted proxies after router creation**

In `cmd/api/main.go`, immediately after `router := gin.New()` (line 166) and before `router.Use(`, insert:

```go
	// Behind Traefik in the container network; trust private ranges so
	// X-Forwarded-For / client IP resolve correctly.
	if err := router.SetTrustedProxies([]string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}); err != nil {
		log.Error("failed to set trusted proxies", "error", err)
		os.Exit(1)
	}
```

- [ ] **Step 3: Build to verify it compiles**

Run: `go build ./cmd/api`
Expected: no output, exit 0.

- [ ] **Step 4: Run the unit suite**

Run: `go test ./internal/... -count=1`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/api/main.go
git commit -m "feat(api): env-driven CORS origins and trusted proxies"
```

---

## Task 3: Rewrite backend Dockerfile (multi-binary, Go 1.25)

**Files:**
- Rewrite: `Dockerfile`

- [ ] **Step 1: Replace `Dockerfile` contents**

```dockerfile
# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /out/worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /out/seed ./cmd/seed

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget && adduser -D -u 10001 app
COPY --from=builder /out/api /usr/local/bin/api
COPY --from=builder /out/worker /usr/local/bin/worker
COPY --from=builder /out/seed /usr/local/bin/seed
COPY migrations /migrations
USER app
EXPOSE 8080
CMD ["api"]
```

- [ ] **Step 2: Build the image**

Run: `docker build -t zoora-backend:test .`
Expected: completes with `naming to docker.io/library/zoora-backend:test`.

- [ ] **Step 3: Verify all three binaries exist and run**

Run: `docker run --rm zoora-backend:test sh -c 'api --help 2>&1 | head -1; ls -1 /usr/local/bin'`
Expected: lists `api`, `seed`, `worker` (the `--help` line may show a config error — that is fine, it proves the binary executes).

- [ ] **Step 4: Commit**

```bash
git add Dockerfile
git commit -m "build(backend): production multi-binary image on Go 1.25"
```

---

## Task 4: Frontend image — nginx static SPA

**Files:**
- Create: `frontend/Dockerfile`
- Create: `frontend/nginx.conf`

- [ ] **Step 1: Create `frontend/nginx.conf`**

```nginx
server {
    listen 80;
    server_name _;
    root /usr/share/nginx/html;
    index index.html;

    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css application/json application/javascript application/xml font/woff2 image/svg+xml;

    # Hashed build assets — cache hard.
    location /assets/ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        try_files $uri =404;
    }

    # SPA fallback — every unknown path serves index.html.
    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

- [ ] **Step 2: Create `frontend/Dockerfile`**

```dockerfile
# syntax=docker/dockerfile:1

FROM node:22-alpine AS builder
WORKDIR /app
RUN corepack enable
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY . .
ARG VITE_API_URL
ENV VITE_API_URL=$VITE_API_URL
RUN pnpm build

FROM nginx:alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
```

- [ ] **Step 3: Build the image (with the build arg)**

Run: `docker build --build-arg VITE_API_URL=https://api.example.com -t zoora-frontend:test ./frontend`
Expected: completes; `pnpm build` runs vite build + `tsc --noEmit` with no errors.

- [ ] **Step 4: Verify static output is present and baked URL is in the bundle**

Run: `docker run --rm zoora-frontend:test sh -c 'test -f /usr/share/nginx/html/index.html && grep -rl "api.example.com" /usr/share/nginx/html/assets | head -1'`
Expected: prints a path under `/usr/share/nginx/html/assets/` (proves `VITE_API_URL` was baked at build time).

- [ ] **Step 5: Commit**

```bash
git add frontend/Dockerfile frontend/nginx.conf
git commit -m "build(frontend): nginx static SPA image with baked API URL"
```

---

## Task 5: LiveKit production config

**Files:**
- Create: `deploy/livekit.yaml`

- [ ] **Step 1: Create `deploy/livekit.yaml`**

Keys are injected at runtime via the `LIVEKIT_KEYS` env (set in compose), so no secrets live in this file.

```yaml
port: 7880
log_level: info
rtc:
  tcp_port: 7881
  udp_port: 7882
  use_external_ip: true
redis:
  address: redis:6379
```

- [ ] **Step 2: Validate it is well-formed YAML**

Run: `docker run --rm -v "$PWD/deploy/livekit.yaml:/etc/livekit.yaml:ro" livekit/livekit-server:latest --config /etc/livekit.yaml --help >/dev/null && echo OK`
Expected: prints `OK` (config file parses; `--help` exits cleanly).

- [ ] **Step 3: Commit**

```bash
git add deploy/livekit.yaml
git commit -m "build(livekit): production server config"
```

---

## Task 6: Production env example + gitignore

**Files:**
- Create: `.env.prod.example`
- Modify: `.gitignore`

- [ ] **Step 1: Create `.env.prod.example`**

```dotenv
# ── Edge / domains ──────────────────────────────────────────────
# Apex domain. Subdomains used: app., api., livekit., s3.
DOMAIN=example.com
# Email for Let's Encrypt registration.
ACME_EMAIL=admin@example.com

# ── App ─────────────────────────────────────────────────────────
ENVIRONMENT=production
PORT=8080
# Browser origin allowed to call the API (the frontend).
CORS_ALLOWED_ORIGINS=https://app.example.com
JWT_SECRET=CHANGE_ME_TO_A_LONG_RANDOM_STRING
JWT_EXPIRY=24h

# ── Postgres (self-hosted container) ────────────────────────────
DB_USERNAME=zoora
DB_PASSWORD=CHANGE_ME
DB_DATABASE=zoora
# Internal connection string used by api/worker/migrate.
DATABASE_URL=postgres://zoora:CHANGE_ME@pgsql:5432/zoora?sslmode=disable

# ── Redis (self-hosted container) ───────────────────────────────
REDIS_URL=redis://redis:6379

# ── LiveKit (self-hosted container) ─────────────────────────────
LIVEKIT_HOST=http://livekit:7880
LIVEKIT_PUBLIC_URL=wss://livekit.example.com
LIVEKIT_API_KEY=CHANGE_ME_KEY
LIVEKIT_API_SECRET=CHANGE_ME_SECRET

# ── S3 / RustFS (self-hosted container) ─────────────────────────
# Public endpoint — presigned URLs are generated against this host.
S3_ENDPOINT=https://s3.example.com
S3_BUCKET=private
S3_ACCESS_KEY=CHANGE_ME
S3_SECRET_KEY=CHANGE_ME
S3_REGION=us-east-1

# ── Frontend build arg (baked at image build time) ──────────────
VITE_API_URL=https://api.example.com
```

- [ ] **Step 2: Ignore the real env file**

Append to `.gitignore` under the `# Environment` section:

```
.env.prod
```

- [ ] **Step 3: Verify the real file would be ignored**

Run: `git check-ignore -v .env.prod`
Expected: prints a line referencing `.gitignore` and `.env.prod` (proves it is ignored).

- [ ] **Step 4: Commit**

```bash
git add .env.prod.example .gitignore
git commit -m "build(deploy): production env example and gitignore"
```

---

## Task 7: Production docker-compose

**Files:**
- Create: `docker-compose.prod.yml`

- [ ] **Step 1: Create `docker-compose.prod.yml`**

```yaml
services:
  traefik:
    image: traefik:v3.1
    restart: unless-stopped
    command:
      - --providers.docker=true
      - --providers.docker.exposedbydefault=false
      - --entrypoints.web.address=:80
      - --entrypoints.websecure.address=:443
      - --entrypoints.web.http.redirections.entrypoint.to=websecure
      - --entrypoints.web.http.redirections.entrypoint.scheme=https
      - --certificatesresolvers.le.acme.email=${ACME_EMAIL}
      - --certificatesresolvers.le.acme.storage=/letsencrypt/acme.json
      - --certificatesresolvers.le.acme.tlschallenge=true
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - "letsencrypt:/letsencrypt"
      - "/var/run/docker.sock:/var/run/docker.sock:ro"
    networks: [zoora]

  frontend:
    build:
      context: ./frontend
      args:
        VITE_API_URL: "https://api.${DOMAIN}"
    image: zoora-frontend:latest
    restart: unless-stopped
    networks: [zoora]
    labels:
      - traefik.enable=true
      - traefik.http.routers.frontend.rule=Host(`app.${DOMAIN}`)
      - traefik.http.routers.frontend.entrypoints=websecure
      - traefik.http.routers.frontend.tls.certresolver=le
      - traefik.http.services.frontend.loadbalancer.server.port=80

  api:
    build:
      context: .
    image: zoora-backend:latest
    command: ["api"]
    restart: unless-stopped
    env_file: [.env.prod]
    environment:
      ENVIRONMENT: production
      DATABASE_URL: "postgres://${DB_USERNAME}:${DB_PASSWORD}@pgsql:5432/${DB_DATABASE}?sslmode=disable"
      REDIS_URL: "redis://redis:6379"
      LIVEKIT_HOST: "http://livekit:7880"
      LIVEKIT_PUBLIC_URL: "wss://livekit.${DOMAIN}"
      S3_ENDPOINT: "https://s3.${DOMAIN}"
      CORS_ALLOWED_ORIGINS: "https://app.${DOMAIN}"
    depends_on:
      pgsql: {condition: service_healthy}
      redis: {condition: service_healthy}
      migrate: {condition: service_completed_successfully}
      bucket-init: {condition: service_completed_successfully}
    networks: [zoora]
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://127.0.0.1:8080/healthz"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 20s
    labels:
      - traefik.enable=true
      - traefik.http.routers.api.rule=Host(`api.${DOMAIN}`)
      - traefik.http.routers.api.entrypoints=websecure
      - traefik.http.routers.api.tls.certresolver=le
      - traefik.http.services.api.loadbalancer.server.port=8080

  worker:
    image: zoora-backend:latest
    command: ["worker"]
    restart: unless-stopped
    env_file: [.env.prod]
    environment:
      ENVIRONMENT: production
      DATABASE_URL: "postgres://${DB_USERNAME}:${DB_PASSWORD}@pgsql:5432/${DB_DATABASE}?sslmode=disable"
      REDIS_URL: "redis://redis:6379"
      LIVEKIT_HOST: "http://livekit:7880"
      LIVEKIT_PUBLIC_URL: "wss://livekit.${DOMAIN}"
      S3_ENDPOINT: "https://s3.${DOMAIN}"
    depends_on:
      pgsql: {condition: service_healthy}
      redis: {condition: service_healthy}
      migrate: {condition: service_completed_successfully}
    networks: [zoora]

  migrate:
    image: migrate/migrate:latest
    restart: "no"
    command:
      - "-path"
      - "/migrations"
      - "-database"
      - "postgres://${DB_USERNAME}:${DB_PASSWORD}@pgsql:5432/${DB_DATABASE}?sslmode=disable"
      - "up"
    volumes:
      - "./migrations:/migrations:ro"
    depends_on:
      pgsql: {condition: service_healthy}
    networks: [zoora]

  bucket-init:
    image: amazon/aws-cli:latest
    restart: "no"
    entrypoint: ["/bin/sh", "-c"]
    command:
      - "aws --endpoint-url http://rustfs:9000 s3 mb s3://${S3_BUCKET} || true"
    environment:
      AWS_ACCESS_KEY_ID: "${S3_ACCESS_KEY}"
      AWS_SECRET_ACCESS_KEY: "${S3_SECRET_KEY}"
      AWS_DEFAULT_REGION: "${S3_REGION:-us-east-1}"
    depends_on:
      rustfs: {condition: service_healthy}
    networks: [zoora]

  seed:
    image: zoora-backend:latest
    command: ["seed"]
    profiles: ["seed"]
    env_file: [.env.prod]
    environment:
      ENVIRONMENT: production
      DATABASE_URL: "postgres://${DB_USERNAME}:${DB_PASSWORD}@pgsql:5432/${DB_DATABASE}?sslmode=disable"
      REDIS_URL: "redis://redis:6379"
      LIVEKIT_HOST: "http://livekit:7880"
      S3_ENDPOINT: "https://s3.${DOMAIN}"
    depends_on:
      migrate: {condition: service_completed_successfully}
    networks: [zoora]

  pgsql:
    image: postgres:18-alpine
    restart: unless-stopped
    environment:
      POSTGRES_DB: "${DB_DATABASE}"
      POSTGRES_USER: "${DB_USERNAME}"
      POSTGRES_PASSWORD: "${DB_PASSWORD}"
    volumes:
      - "pgsql-data:/var/lib/postgresql"
    networks: [zoora]
    healthcheck:
      test: ["CMD", "pg_isready", "-q", "-d", "${DB_DATABASE}", "-U", "${DB_USERNAME}"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:alpine
    restart: unless-stopped
    command: ["redis-server", "--appendonly", "yes"]
    volumes:
      - "redis-data:/data"
    networks: [zoora]
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  rustfs:
    image: rustfs/rustfs:latest
    restart: unless-stopped
    environment:
      RUSTFS_VOLUMES: /data
      RUSTFS_ADDRESS: "0.0.0.0:9000"
      RUSTFS_ACCESS_KEY: "${S3_ACCESS_KEY}"
      RUSTFS_SECRET_KEY: "${S3_SECRET_KEY}"
      RUSTFS_EXTERNAL_ADDRESS: "https://s3.${DOMAIN}"
      RUSTFS_CORS_ALLOWED_ORIGINS: "https://app.${DOMAIN}"
      RUSTFS_LOG_LEVEL: info
    volumes:
      - "rustfs-data:/data"
    networks: [zoora]
    healthcheck:
      test: ["CMD", "sh", "-c", "curl -f http://127.0.0.1:9000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    labels:
      - traefik.enable=true
      - traefik.http.routers.s3.rule=Host(`s3.${DOMAIN}`)
      - traefik.http.routers.s3.entrypoints=websecure
      - traefik.http.routers.s3.tls.certresolver=le
      - traefik.http.services.s3.loadbalancer.server.port=9000

  livekit:
    image: livekit/livekit-server:latest
    restart: unless-stopped
    command: ["--config", "/etc/livekit.yaml"]
    environment:
      LIVEKIT_KEYS: "${LIVEKIT_API_KEY}: ${LIVEKIT_API_SECRET}"
    volumes:
      - "./deploy/livekit.yaml:/etc/livekit.yaml:ro"
    ports:
      - "7881:7881"
      - "7882:7882/udp"
    depends_on:
      redis: {condition: service_healthy}
    networks: [zoora]
    labels:
      - traefik.enable=true
      - traefik.http.routers.livekit.rule=Host(`livekit.${DOMAIN}`)
      - traefik.http.routers.livekit.entrypoints=websecure
      - traefik.http.routers.livekit.tls.certresolver=le
      - traefik.http.services.livekit.loadbalancer.server.port=7880

networks:
  zoora:
    driver: bridge

volumes:
  pgsql-data:
  redis-data:
  rustfs-data:
  letsencrypt:
```

- [ ] **Step 2: Validate compose config resolves**

Run:
```bash
cp .env.prod.example .env.prod
docker compose --env-file .env.prod -f docker-compose.prod.yml config -q && echo CONFIG_OK
rm -f .env.prod
```
Expected: prints `CONFIG_OK` with no errors. (The temp `.env.prod` is removed; it is gitignored anyway.)

- [ ] **Step 3: Commit**

```bash
git add docker-compose.prod.yml
git commit -m "build(deploy): production docker-compose stack with Traefik TLS"
```

---

## Task 8: Deployment runbook

**Files:**
- Create: `docs/DEPLOYMENT.md`

- [ ] **Step 1: Create `docs/DEPLOYMENT.md`**

```markdown
# Production Deployment

Single-VM deployment via `docker-compose.prod.yml`. Traefik terminates TLS and
routes subdomains of one apex. All stateful deps run as containers.

## Prerequisites

- A VM with Docker + Docker Compose plugin.
- A domain you control. Create DNS A records pointing all of these at the VM IP:
  - `app.<domain>`   — frontend
  - `api.<domain>`   — backend API
  - `livekit.<domain>` — LiveKit signaling (wss)
  - `s3.<domain>`    — file storage (presigned URLs)
- Ports open on the VM: `80/tcp`, `443/tcp`, `7881/tcp`, `7882/udp`.

## First deploy

1. Clone the repo onto the VM.
2. Create the env file and fill in real values (strong secrets):
   ```bash
   cp .env.prod.example .env.prod
   # edit .env.prod: DOMAIN, ACME_EMAIL, JWT_SECRET, DB_PASSWORD,
   # LIVEKIT_API_KEY/SECRET, S3_ACCESS_KEY/SECRET, VITE_API_URL=https://api.<domain>
   ```
3. Build and start:
   ```bash
   docker compose --env-file .env.prod -f docker-compose.prod.yml up -d --build
   ```
   On startup: `migrate` runs all migrations, `bucket-init` creates the S3 bucket,
   then `api` and `worker` start. Traefik obtains Let's Encrypt certs automatically
   (allow ~30s on first hit per host).
4. (Optional) Seed baseline data — manual, never automatic:
   ```bash
   docker compose --env-file .env.prod -f docker-compose.prod.yml --profile seed run --rm seed
   ```

## Verify

- `https://api.<domain>/healthz` → `200`.
- `https://app.<domain>` loads the SPA.
- Logs: `docker compose --env-file .env.prod -f docker-compose.prod.yml logs -f api`.

## Updates / redeploy

```bash
git pull
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d --build
```

Migrations re-run idempotently on each deploy.

## Notes

- `VITE_API_URL` is baked into the frontend image at build time. Changing the API
  domain requires a frontend rebuild.
- LiveKit WebRTC media uses `7881/tcp` and `7882/udp` directly on the host (not via
  Traefik). If the VM is behind NAT and auto external-IP detection fails, set the
  host public IP in `deploy/livekit.yaml` under `rtc.node_ip`.
- Backups: snapshot the `pgsql-data` and `rustfs-data` volumes.
```

- [ ] **Step 2: Commit (force-add — `docs/` is gitignored)**

```bash
git add -f docs/DEPLOYMENT.md
git commit -m "docs(deploy): production deployment runbook"
```

---

## Self-Review Notes

- **Spec coverage:** topology/subdomains (Task 7), Traefik TLS (Task 7), backend multi-binary image (Task 3), frontend nginx image + baked `VITE_API_URL` (Task 4), migrate one-shot (Task 7), bucket-init (Task 7), seed profile (Task 7), LiveKit prod config (Task 5), CORS env + trusted proxies (Tasks 1-2), `.env.prod.example` + secrets (Task 6), healthchecks/ordering (Task 7), runbook (Task 8). All spec deliverables mapped.
- **Type consistency:** field `CORSAllowedOrigins` defined in Task 1 is the exact symbol consumed in Task 2. Image tags `zoora-backend:latest` / `zoora-frontend:latest` consistent across compose services. `/healthz` matches the route in `cmd/api/main.go:175`.
- **Verification points to watch during execution:** (a) RustFS sigv4 validation against the forwarded `Host: s3.<domain>` for presigned URLs — confirm a real upload/download round-trips; (b) LiveKit external IP detection on the target VM.
```
