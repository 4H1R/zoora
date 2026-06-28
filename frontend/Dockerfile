FROM node:24-alpine AS builder
WORKDIR /app
RUN corepack enable
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY . .
# Same-origin: the SPA calls /api/v1 on its own host; nginx proxies to the API.
ENV VITE_API_URL=/api/v1
# Tenant config baked at build time. ENV overrides the tracked .env (localhost).
ARG VITE_BASE_DOMAIN
ARG VITE_ADMIN_SUBDOMAIN=admin
ENV VITE_BASE_DOMAIN=${VITE_BASE_DOMAIN}
ENV VITE_ADMIN_SUBDOMAIN=${VITE_ADMIN_SUBDOMAIN}
RUN pnpm build

FROM nginx:alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
