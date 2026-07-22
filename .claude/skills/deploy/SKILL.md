---
name: deploy
description: >
  Deploy the Zoora project to production. cd to the repo root (/home/ghost/web/zoora) and run
  `make prod-up-build`, which rebuilds and starts the prod Docker Compose stack
  (docker-compose.prod.yml + .env.prod). Use when user says "deploy", "ship", "prod up",
  "release", "push to prod", or asks to build and start the production stack.
---

You are deploying Zoora to production via Docker Compose.

`make prod-up-build` runs:
```
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d --build
```
Rebuilds images (`--build`) and starts detached (`-d`).

## Step 1 — Pre-flight checks

Run from repo root `/home/ghost/web/zoora`. Verify prerequisites exist before deploying:

```bash
cd /home/ghost/web/zoora
test -f .env.prod || echo "MISSING .env.prod"
test -f docker-compose.prod.yml || echo "MISSING docker-compose.prod.yml"
docker info >/dev/null 2>&1 || echo "Docker daemon not running"
```

If `.env.prod` or `docker-compose.prod.yml` is missing, or Docker is down, stop and tell the user — do not proceed.

## Step 2 — Confirm intent

Deploying to production is outward-facing and hard to reverse. Unless the user already explicitly said to deploy/ship to prod in this turn, confirm before running.

## Step 3 — Deploy

```bash
cd /home/ghost/web/zoora && make prod-up-build
```

Build can take several minutes. Do not interrupt it.

## Step 4 — Verify

Check containers came up healthy:

```bash
cd /home/ghost/web/zoora && make prod-ps
```

If any service is not `Up`/healthy, inspect logs:

```bash
cd /home/ghost/web/zoora && make prod-logs
```

Report which services are running and flag anything crash-looping or exited.

## Related targets

- `make prod-up` — start without rebuild
- `make prod-down` — stop the stack
- `make prod-down-purge` — stop + remove volumes (destructive, wipes data)
- `make prod-logs` — follow logs
- `make prod-ps` — list container status
- `make prod-seed` — seed prod DB

## Report

State the outcome plainly: build succeeded/failed, which services are up, any errors from logs. If the build or a container failed, quote the exact error.
