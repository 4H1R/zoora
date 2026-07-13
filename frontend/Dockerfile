FROM oven/bun:1-alpine AS builder
WORKDIR /app

ENV HUSKY=0

COPY package.json bun.lock bunfig.toml ./
COPY frontend/package.json ./frontend/package.json
COPY frontend/patches/ ./frontend/patches/
RUN bun install --frozen-lockfile

COPY frontend/ ./frontend/

ENV VITE_API_URL=/api/v1
ENV VITE_WS_URL=/api/v1/ws
ARG VITE_BASE_DOMAIN
ARG VITE_ADMIN_SUBDOMAIN=admin
ENV VITE_BASE_DOMAIN=${VITE_BASE_DOMAIN}
ENV VITE_ADMIN_SUBDOMAIN=${VITE_ADMIN_SUBDOMAIN}

RUN bun --filter frontend build

FROM oven/bun:1-alpine AS runtime
WORKDIR /app

ENV NODE_ENV=production
COPY --from=builder /app/frontend/.output ./.output
ENV PORT=3000
EXPOSE 3000

CMD ["bun", "run", ".output/server/index.mjs"]
