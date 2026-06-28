# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod go.sum ./
RUN go mod download
# docs/ is gitignored (swag-generated). Install swag as a standalone tool
# (its CLI deps aren't in the project go.sum) so the build is self-contained.
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6
COPY . .
# Regenerate the swagger docs package the api binary imports.
RUN swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
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
