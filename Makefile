EXEC := docker compose exec -T app
EXEC_TTY := docker compose exec app
PROD := docker compose --env-file .env.prod -f docker-compose.prod.yml

.PHONY: up down build restart ps logs run-api run-worker migrate-up migrate-reset migrate-create test test-integration lint swagger seed shell prod

up:
	docker compose up -d

down:
	docker compose down

build:
	docker compose build

restart:
	docker compose restart

ps:
	docker compose ps

logs:
	docker compose logs -f

run-api:
	$(EXEC) go run cmd/api/main.go

run-worker:
	$(EXEC) go run cmd/worker/main.go

migrate-up:
	$(EXEC) sh -c 'migrate -path migrations -database "$$DATABASE_URL" up'

migrate-reset:
	$(EXEC) sh -c 'migrate -path migrations -database "$$DATABASE_URL" drop -f'
	$(EXEC) sh -c 'migrate -path migrations -database "$$DATABASE_URL" up'

migrate-create:
	$(EXEC) migrate create -ext sql -dir migrations -seq $(name)
	$(EXEC) rm -f migrations/*_$(name).down.sql

test:
	$(EXEC) go test ./internal/... -v -count=1

test-integration:
	$(EXEC) go test -tags=integration ./tests/... -v -count=1

lint:
	$(EXEC) sh -c 'GOFLAGS=-buildvcs=false golangci-lint run ./...'

swagger:
	$(EXEC) swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

generate: swagger
	cd frontend && pnpm generate

seed:
	$(EXEC_TTY) go run cmd/seed/main.go

shell:
	$(EXEC_TTY) bash

prod:
	$(PROD) up -d --build
