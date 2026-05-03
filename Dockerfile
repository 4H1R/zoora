FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /worker ./cmd/worker

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /api /usr/local/bin/api
COPY --from=builder /worker /usr/local/bin/worker
COPY migrations /migrations

EXPOSE 8080

CMD ["api"]
