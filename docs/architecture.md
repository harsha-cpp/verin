# Architecture Notes

## Runtime shape

- `web` is a responsive SPA with a quiet app shell, command surface, and core operational screens for documents, search, audit, and admin.
- `api` exposes REST endpoints under `/api/v1`, owns auth/session state, validates permissions, creates signed URLs, and persists the system of record.
- `worker` consumes Asynq jobs for OCR extraction, preview registration, in-app notifications, audit export generation, and retention hooks.

## Data flow

1. Client requests `init-upload`.
2. API creates an upload row and returns a signed object-storage URL.
3. Client uploads directly to MinIO/S3-compatible storage.
4. Client calls `complete-upload`.
5. API creates document/version records, persists metadata/tags, logs audit state, and queues OCR + preview jobs.
6. Worker extracts text through Tika, registers preview assets, and marks document processing state complete.

## Access model

- Product behavior is single-org for the MVP, but org identifiers remain on tenant-owned tables.
- Auth uses app-managed secure cookies with Redis-backed session state and TOTP MFA.
- Authorization combines:
  - global role grants
  - collection membership
  - document-level overrides
  - owner/admin checks

## Storage and search

- Binary data never enters Postgres.
- Object keys follow:
  - `orgs/{orgId}/documents/{documentId}/versions/{versionId}/original`
  - `orgs/{orgId}/documents/{documentId}/versions/{versionId}/preview`
  - `orgs/{orgId}/exports/{exportId}.csv`
- Search is Postgres-first:
  - document title/filename vector
  - OCR text vector
  - metadata text filtering
  - GIN + trigram indexes

## Operational notes

- Local infrastructure is provided through `docker-compose.yml` with Postgres, Redis, MinIO, Tika, and Mailpit.
- OpenAPI is the contract source of truth for the generated TS client.
- The current frontend bundle is intentionally feature-complete before aggressive code-splitting; Vite warns about chunk size and that should be addressed in a later performance pass.
