# Solo Build Prompt — DMS (Verin)

Build a production-grade MVP for a cloud-native Document Manager System (DMS).

Core constraints:
- absolutely no AI-slop
- UI must feel like a top SaaS product
- clean, minimal, premium, quiet, sharp
- no generic dashboard template look
- use a custom font, preferably Geist
- prioritize clarity, spacing, typography, and accuracy
- frontend must be React + TypeScript
- backend must be Go
- database must be PostgreSQL
- file storage must use S3-compatible object storage
- architecture must support async jobs for OCR, previews, notifications, and retention tasks
- design must be API-first and production-minded

Build the system with these modules:

Frontend:
- React + TypeScript + Vite
- Tailwind CSS
- shadcn/ui + Radix UI
- TanStack Query
- TanStack Table
- React Hook Form + Zod
- command bar/global search
- responsive app shell
- premium document list, document details, upload flow, search UI, audit UI, admin UI
- excellent empty/loading/error states
- keyboard-friendly interactions
- WCAG-conscious accessibility
- custom font integration
- visually calm, minimal, polished styling

Backend:
- Go
- Chi router
- pgx + sqlc
- structured logging
- OpenTelemetry hooks
- REST API under /api/v1
- OpenAPI spec
- auth, MFA, RBAC, documents, versions, metadata, search, audit, comments, notifications, admin, quotas, retention
- signed upload flow to object storage
- strict permission middleware
- consistent error envelopes
- background job support

Database:
- PostgreSQL schema for users, roles, user_roles, documents, document_versions, metadata, tags, collections, permissions, OCR text, audit_events, comments, notifications, saved_searches, quotas, retention_policies, uploads
- proper indexes
- migration files
- seed data for dev

Storage:
- S3-compatible abstraction
- signed upload/download URLs
- clear key naming
- support previews and versioned binaries

Async processing:
- worker service
- OCR pipeline
- preview generation
- retention/archive tasks
- notification dispatch
- export generation

Search:
- metadata search
- full-text OCR search
- Postgres FTS + GIN indexes
- saved searches
- filter chips and advanced filtering in UI

Security:
- TLS-aware architecture
- bcrypt or argon2id password hashing
- TOTP MFA
- rate limiting
- signed URL expiry
- secure file validation
- audit logging of sensitive actions
- prepared SQL only
- role and resource-level authorization

Admin:
- users
- roles
- quotas
- retention
- jobs
- health
- usage summaries

Required screens:
- login
- MFA setup/verify
- dashboard
- documents list
- document details
- upload
- search
- version history
- audit logs
- admin users/roles
- admin quotas/retention/settings

Implementation requirements:
- create a clean monorepo or clearly structured multi-app setup
- include README with local setup
- include .env.example files
- include migration scripts
- include sample seed data
- include OpenAPI documentation
- include concise architecture notes
- keep code modular and readable
- avoid over-abstraction
- keep naming exact and professional

Most important:
- no UI slop
- no fake complexity
- no unnecessary animation
- no generic component spam
- no hand-wavy architecture
- make the product feel intentional and premium

At the end of the implementation, ask me exactly which environment variables and deployment target I want to use before finalizing deployment-specific configuration.

Required final question:
“Tell me what environment variables you want and which deployment target you want to use, and I’ll wire the final deployment configuration.”
