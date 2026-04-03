SHELL := /bin/zsh

.PHONY: install dev infra-up infra-down build test lint check generate migrate seed api worker

install:
	pnpm install
	cd apps/backend && go mod tidy

dev:
	pnpm dev

infra-up:
	docker compose up -d

infra-down:
	docker compose down --remove-orphans

build:
	pnpm build
	cd apps/backend && go build ./...

test:
	pnpm test
	cd apps/backend && go test ./...

lint:
	pnpm lint
	cd apps/backend && go test ./... 

check:
	pnpm check
	cd apps/backend && go test ./...

generate:
	pnpm generate:api-client
	cd apps/backend && sqlc generate

migrate:
	psql $$DATABASE_URL -f infra/migrations/001_initial.sql

seed:
	psql $$DATABASE_URL -f infra/seed/dev_seed.sql

api:
	cd apps/backend && go run ./cmd/api

worker:
	cd apps/backend && go run ./cmd/worker
