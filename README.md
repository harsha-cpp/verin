# Verin DMS MVP

Production-minded MVP for a cloud-native document management system with a React frontend, Go API/worker, PostgreSQL, Redis-backed jobs, and S3-compatible storage.

## Stack

- `apps/web`: React, TypeScript, Vite, Tailwind CSS, TanStack Query/Table, RHF + Zod
- `apps/backend`: Go, Chi, pgx, sqlc, Asynq, Redis, MinIO-compatible storage
- `packages/ui`: shared UI primitives
- `packages/api-client`: generated OpenAPI client wrapper
- `infra`: SQL migrations, seed data, local infrastructure
- `docs`: OpenAPI, architecture notes, API notes

## Local setup

1. Copy environment templates:
   - `cp apps/backend/.env.example apps/backend/.env`
   - `cp apps/web/.env.example apps/web/.env`
2. Start local services:
   - `docker compose up -d`
3. Install dependencies:
   - `pnpm install`
   - `cd apps/backend && go mod tidy`
4. Apply schema and seed data:
   - `export DATABASE_URL=postgres://verin:verin@localhost:5432/verin?sslmode=disable`
   - `make migrate`
   - `make seed`
5. Generate the API client if needed:
   - `pnpm generate:api-client`
6. Run the services:
   - `make api`
   - `make worker`
   - `pnpm --filter @verin/web dev`

## Seeded local accounts

- `admin@verin.local`
- `editor@verin.local`
- `auditor@verin.local`
- Password for all three: `verin123!`

## Common commands

- `pnpm build`
- `pnpm test`
- `cd apps/backend && go test ./...`
- `pnpm generate:api-client`
- `~/go/bin/sqlc generate`

## Notes

- Uploads use signed direct-to-storage URLs and are finalized through `/api/v1/documents/complete-upload`.
- OCR, preview generation, notifications, and audit exports are queued through Asynq and handled by the Go worker.
- The frontend includes premium fallback data so the UI remains inspectable even if the API is not running.
- Deployment-specific configuration is intentionally deferred.
# verin
