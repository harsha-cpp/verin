# Document Manager System (DMS) — Product PRD + AI Build Document

> Implementation-grade product and engineering brief derived from the uploaded SRS, adapted to the requested stack and product direction: **React frontend, Go backend, PostgreSQL, object storage, premium SaaS-grade UI, custom font, clean/simple/accurate output, no AI-slop.** Source: uploaded SRS for the DMS. fileciteturn0file0

---

## 1) Product Overview

Build a **cloud-native Document Manager System** for secure document upload, storage, versioning, metadata management, OCR-powered search, audit trails, collaboration, and admin visibility.

This product is **not** a document editor. It is a **document operations platform**.

### Primary outcomes
- users upload and organize documents with confidence
- users retrieve documents fast
- admins control access precisely
- auditors can verify activity and export reports
- the system scales cleanly without UI clutter or backend chaos

### Product feel
The UI must feel like a top SaaS product:
- calm
- sharp
- minimal
- premium typography
- no wasted visual noise
- no generic dashboard-template look
- no “AI generated” bloat

---

## 2) Scope

### In scope
- authentication
- MFA
- RBAC
- SSO-ready architecture
- document upload and processing
- object storage integration
- metadata tagging
- folder/collection organization
- versioning and rollback
- preview and download
- OCR extraction
- full-text and metadata search
- audit logs and compliance reporting
- collaboration basics
- email notifications
- admin dashboard
- health/monitoring endpoints
- backup/restore of metadata
- API-first platform
- responsive SPA

### Out of scope for v1
- full document authoring
- Google Docs style real-time co-editing
- e-signature workflows
- complex BPM/workflow builder
- bulk paper-to-digital migration ops tooling
- native mobile app
- advanced AI assistant features inside the product

---

## 3) Goals

### Business goals
- reduce time to find documents
- improve operational control
- support compliance and auditability
- enable structured document lifecycle management
- create a trustworthy internal system of record

### User goals
- upload without friction
- search in seconds
- understand permissions clearly
- review changes/version history easily
- know whether a document is current, archived, or restricted

### Success metrics
- median search response under 500 ms for indexed queries
- upload reliability above 99.5%
- first document retrieval in under 3 clicks for common flows
- near-zero permission ambiguity in user testing
- audit export flow completed in under 2 minutes
- admin setup for a new team in under 15 minutes

---

## 4) Users and Roles

### General User
Can upload, tag, search, preview, and download allowed documents.

### Power User
Can manage structure, metadata quality, batch operations, saved filters, and version review.

### Admin
Can manage users, roles, quotas, retention settings, permissions, jobs, and system-level views.

### Compliance Auditor
Can access audit logs, generate compliance reports, inspect access history, and verify integrity controls.

---

## 5) Product Principles

### UX principles
- one strong primary action per screen
- everything obvious at a glance
- advanced controls hidden until needed
- empty states teach the workflow
- consistent interaction patterns
- fast keyboard-first search and navigation
- dense where needed, airy where useful
- never decorate where clarity should lead

### Engineering principles
- API first
- typed contracts
- secure by default
- async processing for heavy jobs
- explicit permissions model
- observable from day one
- modular codebase, not accidental spaghetti
- deployment-agnostic enough to finalize later

---

## 6) UX / UI Direction

## 6.1 Design language
Target a look that combines:
- Linear restraint
- Notion calm
- Vercel polish
- enterprise trust without enterprise ugliness

### Visual characteristics
- neutral palette
- restrained accent color
- subtle borders
- minimal shadows
- almost no gradients
- excellent spacing rhythm
- large readable headings
- crisp table design
- simple icons
- clean command surfaces
- no oversized cards unless necessary

## 6.2 Font system
Use a custom font with premium feel.

### Recommended
- **Geist** as primary UI font
- fallback: `Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`

### Optional pairing
- UI/body: Geist
- mono: JetBrains Mono or Geist Mono for hashes, IDs, diffs, logs

## 6.3 Component styling rules
- radius: 10–14px
- border-first surfaces
- soft hover states
- strong focus rings
- line height generous but not loose
- tables compact and readable
- forms narrow and disciplined
- only one accent color in primary flows

## 6.4 Frontend stack
- React
- TypeScript
- Vite
- Tailwind CSS
- shadcn/ui
- Radix UI
- TanStack Query
- TanStack Table
- React Hook Form
- Zod
- Framer Motion only for subtle transitions
- Lucide icons

This stack gives full control and avoids template feel.

---

## 7) Information Architecture

## 7.1 Main app areas
- Login
- MFA setup / verification
- Dashboard
- Documents
- Search
- Upload
- Collections / Folders
- Shared with me
- Audit
- Reports
- Admin
- Settings
- Profile

## 7.2 Primary navigation
Left sidebar, very quiet:
- Dashboard
- Documents
- Search
- Upload
- Shared
- Audit
- Admin
- Settings

Topbar:
- global search / command bar
- notifications
- quick upload
- profile menu

## 7.3 Key screens
1. Dashboard
2. Document list
3. Document details
4. Upload modal / upload page
5. Search results
6. Version history
7. Audit log explorer
8. Admin users and roles
9. Admin quotas and retention
10. Report export flow

---

## 8) Feature Requirements

## 8.1 Authentication and Access Control

### Functional requirements
- email/password auth
- TOTP MFA
- session timeout
- account lockout after repeated failed attempts
- RBAC by role
- resource-level access overrides
- admin can assign and revoke roles
- SSO-ready interfaces for later LDAP/AD/OIDC/SAML integration
- optional invite-only organization onboarding

### Roles
- `general_user`
- `power_user`
- `admin`
- `compliance_auditor`

### Permission groups
- document.read
- document.write
- document.delete
- document.share
- document.restore
- metadata.manage
- audit.read
- report.export
- user.manage
- role.manage
- quota.manage
- retention.manage
- system.settings.read
- system.settings.write

### Non-functional
- all auth over TLS
- secure cookie or token strategy
- CSRF-safe session flow if cookie auth used
- 15-minute inactivity timeout for sensitive contexts
- lockout after repeated failures
- audit all auth events

---

## 8.2 Document Upload and Lifecycle

### Functional requirements
- drag-and-drop upload
- manual file picker
- progress indicator
- resumable uploads
- checksum generation
- MIME/type validation
- file size validation
- processing state tracking
- archive policy handling
- soft delete
- hard delete with authorization gates
- upload batch support
- metadata entry during or after upload

### Supported v1 file types
- PDF
- DOCX
- PNG
- JPG/JPEG

### Lifecycle states
- `uploaded`
- `processing`
- `ready`
- `archived`
- `deleted`
- `processing_failed`

### Non-functional
- async processing only
- UI never blocked by OCR or preview generation
- idempotent upload finalization
- checksum-based integrity verification
- clean retry flow on failure

---

## 8.3 Metadata Management

### System-generated metadata
- document ID
- original filename
- storage key
- size
- mime type
- checksum
- uploader
- created at
- updated at
- latest version
- status

### User-managed metadata
- title
- description
- tags
- category
- owner
- department
- expiry date
- reference number
- custom fields

### Rules
- metadata schema must be extensible
- admins can add organization-specific custom fields
- fields can be typed: text, date, enum, boolean, number
- required fields configurable by category or collection

---

## 8.4 Versioning and Rollback

### Functional requirements
- every meaningful replacement/edit creates a version
- keep full version history
- diff support where format allows
- preview older versions
- restore previous version
- retention cap configurable
- old versions purge policy configurable

### Non-functional
- version storage efficient
- rollback audited
- current version clearly marked
- destructive restore actions require confirmation

---

## 8.5 Search and Retrieval

### Functional requirements
- global search
- keyword search
- full-text OCR search
- metadata filters
- faceted filters
- date range filters
- saved searches
- relevance ranking
- folder / owner / tag filters
- recent searches

### Search experience requirements
- command bar searchable from anywhere
- result grouping by best match / recent / type
- filter chips visible and removable
- result preview snippet where possible
- search state preserved when navigating back

### Non-functional
- indexed metadata in Postgres
- full-text strategy via Postgres FTS or external search later
- OCR text searchable after processing completes
- target sub-500 ms for common indexed queries

---

## 8.6 Audit Trails and Compliance Reporting

### Functional requirements
- log all key actions:
  - login
  - logout
  - upload
  - download
  - preview
  - edit metadata
  - delete
  - restore
  - permission changes
  - admin actions
- filter audit logs by user/date/action/document
- export CSV/PDF reports
- compliance-friendly immutable-style event design
- suspicious activity flagging hooks

### Audit event fields
- event ID
- actor user ID
- actor role
- action type
- resource type
- resource ID
- timestamp
- IP
- user agent
- request ID
- metadata diff where applicable

---

## 8.7 Collaboration and Notifications

### Functional requirements
- shared folders / collections
- document share with inherited RBAC
- document comments
- mentions
- expiry alerts
- share notifications
- archive / retention notifications
- admin alerts for suspicious activity

### Notification channels
- in-app
- email

### Future-safe
- webhook events
- Slack/email provider abstraction
- event-driven notification service

---

## 8.8 Administration and Monitoring

### Functional requirements
- user management
- role assignment
- quota configuration
- retention policy config
- storage usage overview
- processing job status
- failed job retry
- system health view
- backup/restore actions for metadata
- API key or integration settings later

### Health metrics
- app uptime
- DB latency
- queue depth
- OCR job failures
- storage errors
- upload success rate
- search latency
- email delivery failures

---

## 9) Detailed UX Requirements

## 9.1 Dashboard
Purpose:
- immediate clarity
- recent documents
- pending processing
- quick upload
- saved searches
- notifications

Widgets:
- recent activity
- recently viewed
- processing queue summary
- storage usage summary
- quick actions

Do not overload this screen. It must remain calm.

## 9.2 Document List
Requirements:
- clean table/grid toggle
- sortable columns
- bulk select
- filters
- status badges
- compact metadata visibility
- keyboard navigation
- sticky search/filter bar

Columns:
- name
- type
- owner
- tags
- updated at
- version
- status
- access

## 9.3 Document Details
Must show:
- title
- preview
- metadata
- activity
- version timeline
- permissions
- related documents
- comments
- actions

Actions:
- download
- share
- edit metadata
- upload new version
- archive
- delete
- restore old version

## 9.4 Upload Flow
Preferred experience:
1. open upload drawer/modal
2. drop files
3. show upload progress
4. attach metadata
5. assign collection/folder
6. finish and background process

Need excellent feedback states:
- uploading
- processing
- ready
- failed

## 9.5 Search
Search must be one of the best parts of the product.

Requirements:
- command bar
- dedicated search page
- full-text + metadata
- snippets
- filter chips
- saved searches
- recent searches
- keyboard friendly

## 9.6 Admin
Admin UX must feel premium, not legacy.
Use:
- segmented tabs
- calm tables
- compact summaries
- clear destructive warnings
- diff views for permission changes

---

## 10) Frontend Architecture

## 10.1 Suggested app structure
```txt
apps/web
  src/
    app/
    components/
      ui/
      layout/
      documents/
      search/
      upload/
      audit/
      admin/
    features/
      auth/
      documents/
      search/
      versions/
      audit/
      admin/
      notifications/
    hooks/
    lib/
      api/
      auth/
      utils/
      config/
    routes/
    store/
    styles/
    types/
```

## 10.2 Frontend choices
- **React + TypeScript**
- **Vite**
- **TanStack Router** or React Router
- **TanStack Query** for server state
- **Zustand** only if lightweight client state is needed
- **React Hook Form + Zod**
- **Tailwind + shadcn/ui**

## 10.3 Frontend patterns
- feature-oriented modules
- typed API client generated from OpenAPI if possible
- optimistic UI only where safe
- server state and form state kept separate
- error boundaries at route and major panel level
- skeletons instead of spinners where possible

## 10.4 Accessibility
- WCAG 2.1 AA baseline
- keyboard navigation everywhere
- focus-visible states
- proper aria labeling
- table semantics
- accessible dialogs and command palette

---

## 11) Backend Architecture

## 11.1 Backend stack
- **Go**
- recommended HTTP framework: **Gin**, **Chi**, or **Fiber**
- recommended choice: **Chi** for clean, minimal, production-friendly routing
- ORM/query layer:
  - **sqlc** + pgx preferred for type-safe SQL and control
  - or **GORM** if speed of iteration matters more than query clarity
- queue:
  - **Asynq** with Redis, or **River** with Postgres
- auth:
  - session or token-based auth
  - TOTP support
- API:
  - REST first
  - OpenAPI documented
- object storage:
  - S3-compatible abstraction
- observability:
  - OpenTelemetry
  - structured logs
  - Prometheus metrics

### Recommended backend decision
Use:
- **Chi**
- **pgx**
- **sqlc**
- **Redis + Asynq**
- **S3-compatible storage**
- **OpenTelemetry**
- **Zap or Zerolog**

This gives clean control and avoids heavy abstraction.

## 11.2 Suggested services/modules
```txt
apps/api
  cmd/api
  cmd/worker
  internal/
    auth/
    users/
    roles/
    permissions/
    documents/
    uploads/
    versions/
    metadata/
    search/
    audit/
    notifications/
    admin/
    jobs/
    storage/
    ocr/
    previews/
    reports/
    config/
    db/
    http/
    middleware/
    observability/
```

## 11.3 Backend responsibilities
### API service
- auth
- REST endpoints
- permission checks
- metadata CRUD
- search queries
- report generation requests

### Worker service
- OCR jobs
- thumbnail/preview generation
- metadata extraction
- retention/archive jobs
- notification dispatch
- audit export generation

### Scheduler/cron
- archive inactive documents
- purge expired versions
- retention checks
- cleanup orphan uploads
- periodic health summaries

---

## 12) API Design

## 12.1 API style
- REST JSON
- versioned under `/api/v1`
- strong request/response contracts
- OpenAPI spec required
- request IDs on all responses
- consistent error envelopes

## 12.2 Example endpoint groups
### Auth
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/mfa/verify`
- `POST /api/v1/auth/mfa/setup`
- `GET /api/v1/auth/me`

### Users / roles
- `GET /api/v1/users`
- `POST /api/v1/users`
- `PATCH /api/v1/users/:id`
- `GET /api/v1/roles`
- `POST /api/v1/users/:id/roles`

### Documents
- `POST /api/v1/documents/init-upload`
- `POST /api/v1/documents/complete-upload`
- `GET /api/v1/documents`
- `GET /api/v1/documents/:id`
- `PATCH /api/v1/documents/:id`
- `DELETE /api/v1/documents/:id`
- `POST /api/v1/documents/:id/archive`
- `POST /api/v1/documents/:id/restore`

### Versions
- `GET /api/v1/documents/:id/versions`
- `POST /api/v1/documents/:id/versions`
- `POST /api/v1/documents/:id/versions/:versionId/restore`

### Search
- `GET /api/v1/search?q=...`
- `POST /api/v1/search/advanced`
- `POST /api/v1/search/saved`
- `GET /api/v1/search/saved`

### Audit
- `GET /api/v1/audit/events`
- `POST /api/v1/audit/reports/export`

### Admin
- `GET /api/v1/admin/usage`
- `GET /api/v1/admin/jobs`
- `POST /api/v1/admin/jobs/:id/retry`
- `GET /api/v1/admin/health`

## 12.3 Error format
```json
{
  "error": {
    "code": "DOCUMENT_NOT_FOUND",
    "message": "Document not found",
    "details": {}
  },
  "requestId": "req_123"
}
```

---

## 13) Data Architecture

## 13.1 Database
Use **PostgreSQL** as the system of record for:
- users
- roles
- permissions
- document metadata
- versions metadata
- OCR text references or indexed content markers
- audit events
- comments
- notifications
- quotas
- retention policies
- jobs metadata
- saved searches

## 13.2 Storage
Use **S3-compatible object storage** for:
- original files
- version binaries
- preview assets
- thumbnails
- exported reports
- temporary processing artifacts if needed

Store binaries in object storage, never in Postgres.

## 13.3 Search strategy
### V1 recommended
- Postgres full-text search for metadata + OCR text
- GIN indexes
- trigram where useful
- upgrade path to OpenSearch/Meilisearch/Typesense later if needed

This keeps v1 lean and good enough.

---

## 14) Suggested Database Schema

## 14.1 Core tables
- `users`
- `roles`
- `permissions`
- `user_roles`
- `documents`
- `document_versions`
- `document_metadata`
- `document_tags`
- `tags`
- `collections`
- `collection_members`
- `document_permissions`
- `uploads`
- `ocr_jobs`
- `ocr_text`
- `preview_assets`
- `audit_events`
- `comments`
- `notifications`
- `saved_searches`
- `retention_policies`
- `quotas`
- `system_settings`

## 14.2 Important table notes

### users
Fields:
- id
- email
- password_hash
- mfa_enabled
- mfa_secret_encrypted
- status
- created_at
- updated_at
- last_login_at

### documents
Fields:
- id
- org_id
- title
- original_filename
- mime_type
- size_bytes
- checksum_sha256
- owner_user_id
- collection_id
- current_version_id
- status
- archived_at
- deleted_at
- created_at
- updated_at

### document_versions
Fields:
- id
- document_id
- version_number
- storage_key
- size_bytes
- checksum_sha256
- uploaded_by
- created_at
- change_summary

### document_metadata
Fields:
- id
- document_id
- schema_key
- value_text
- value_number
- value_boolean
- value_date
- value_jsonb

### ocr_text
Fields:
- id
- document_version_id
- content
- extraction_status
- extracted_at
- language
- confidence_score

### audit_events
Fields:
- id
- actor_user_id
- actor_role
- action
- resource_type
- resource_id
- ip_address
- user_agent
- request_id
- payload_json
- created_at

---

## 15) Search and OCR Design

## 15.1 OCR pipeline
1. upload finalized
2. worker picks job
3. extract text
4. persist OCR text
5. update search indexable record
6. generate preview
7. mark processing complete

## 15.2 OCR provider options
### Self-hosted/open source
- Tesseract for basic OCR
- OCRmyPDF for PDFs
- Apache Tika for extraction support

### Managed later
- AWS Textract
- Google Document AI
- Azure AI Document Intelligence

### Recommended v1 approach
Start with:
- Apache Tika for extraction where possible
- OCRmyPDF / Tesseract for scanned PDFs/images
- keep provider interface abstract

---

## 16) Storage Design

## 16.1 Storage abstraction
Define a storage interface in Go:
- PutObject
- GetObject
- DeleteObject
- HeadObject
- CreateSignedUploadURL
- CreateSignedDownloadURL
- CopyObject

Support:
- AWS S3
- Cloudflare R2
- MinIO
- DigitalOcean Spaces
- GCS via compatibility layer later if needed

## 16.2 Object key pattern
```txt
orgs/{orgId}/documents/{documentId}/versions/{versionId}/original
orgs/{orgId}/documents/{documentId}/versions/{versionId}/preview
orgs/{orgId}/documents/{documentId}/versions/{versionId}/thumb
orgs/{orgId}/exports/{exportId}
```

## 16.3 Upload flow
Preferred:
- backend creates signed upload URL
- frontend uploads directly to storage
- backend finalizes upload and creates processing job

This reduces API bandwidth pressure.

---

## 17) Security Design

## 17.1 Core requirements
- TLS everywhere
- encryption at rest for sensitive secrets
- strong password hashing
- MFA support
- least privilege RBAC
- signed URLs with short TTL
- audit logging
- rate limiting
- prepared SQL only
- secure file validation
- MIME sniffing protection
- restricted download permissions
- secret rotation support

## 17.2 Security controls
- bcrypt or argon2id for password hashing
- encrypted MFA secrets
- Redis-backed or DB-backed session invalidation
- IP/user rate limiting
- signed URL expiration
- antivirus scanning hook
- file type allowlist
- structured security event logs
- admin actions require elevated checks

## 17.3 Compliance-minded controls
- right to export/delete user-linked data where policy allows
- append-only style audit record pattern
- long retention for audit logs
- explicit destructive action approvals
- permission change logging

---

## 18) Notification Design

## 18.1 Event types
- document shared
- mention in comment
- upload processing finished
- upload failed
- archive warning
- retention expiry
- suspicious access alert
- admin role granted/revoked

## 18.2 Notification system
Recommended:
- event-driven internal notification service
- email provider abstraction
- templates stored in code or DB
- in-app notifications persisted in Postgres

### Email providers
Keep abstraction for later selection:
- Resend
- SendGrid
- SES
- Postmark

---

## 19) Observability and Operations

## 19.1 Logging
- structured JSON logs
- request IDs
- job IDs
- user IDs where appropriate
- avoid leaking secrets

## 19.2 Metrics
- request count
- error rate
- latency
- upload success rate
- OCR processing time
- queue depth
- DB query latency
- search latency
- signed URL generation failures

## 19.3 Tracing
- OpenTelemetry for API and worker
- trace upload finalization, OCR, preview generation, export generation

## 19.4 Monitoring
- Prometheus metrics endpoint
- Grafana dashboards later
- health endpoints for app, DB, queue, storage dependency

---

## 20) Suggested Monorepo / Project Structure

```txt
dms/
  apps/
    web/
    api/
    worker/
  packages/
    ui/
    config/
    types/
    api-client/
  infra/
    docker/
    migrations/
    seed/
  docs/
    prd.md
    api.md
    architecture.md
    deployment.md
```

Alternative:
- keep frontend and backend separate repos if team prefers
- monorepo is cleaner for shared types and coordinated delivery

---

## 21) Recommended Tech Stack Summary

## Frontend
- React
- TypeScript
- Vite
- Tailwind CSS
- shadcn/ui
- Radix UI
- TanStack Query
- TanStack Table
- React Hook Form
- Zod
- Lucide

## Backend
- Go
- Chi
- pgx
- sqlc
- Redis
- Asynq
- OpenTelemetry
- Zap/Zerolog

## Database
- PostgreSQL

## Storage
- S3-compatible object storage

## Auth
- app-managed auth + TOTP MFA
- SSO-ready architecture for later

## Search/OCR
- Postgres FTS
- Apache Tika
- OCRmyPDF / Tesseract
- abstraction for managed OCR later

## Infra basics
- Docker
- reverse proxy
- worker process
- migrations
- env-based config
- object storage bucket(s)
- Redis
- monitoring hooks

---

## 22) Delivery Plan

## Phase 1 — Foundation
- repo setup
- frontend shell
- backend service skeleton
- auth
- RBAC
- Postgres schema
- object storage integration
- signed upload flow

## Phase 2 — Core Documents
- document CRUD
- metadata
- collections
- upload states
- preview/download
- admin user management

## Phase 3 — Processing and Search
- OCR jobs
- full-text indexing
- search UI
- filters
- saved searches
- versioning

## Phase 4 — Audit and Admin
- audit explorer
- report exports
- quotas
- retention
- monitoring surfaces

## Phase 5 — Polish
- UX refinement
- keyboard flows
- accessibility pass
- performance pass
- empty/loading/error states
- premium visual cleanup

---

## 23) Acceptance Criteria

The product is ready for initial release when:
- auth + MFA works
- role-based permissions work reliably
- uploads are stable
- documents store in object storage
- metadata persists in Postgres
- OCR pipeline runs asynchronously
- search returns metadata and OCR matches
- version history is visible and restorable
- audit logs are exportable
- admin can manage users/roles/quotas
- UI is clean, responsive, and clearly premium
- observability is in place
- API documented in OpenAPI

---

## 24) Build Constraints for the Team / AI Builder

- do not generate bloated abstractions
- do not use generic dashboard templates
- do not add random gradients, oversized cards, or noisy charts
- do not over-engineer microservices in v1
- do not couple binary storage to Postgres
- do not block UI on OCR or preview generation
- do not build “AI features” unless explicitly requested
- keep naming clean and precise
- keep components small
- keep SQL explicit
- keep UX ruthless and simple

---

## 25) Solo Build Prompt

Use this prompt as the master build prompt for an AI coding agent or solo implementation workflow.

```md
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
- include Docker setup for local dev
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
```

---

## 26) Final Note for Implementation Flow

Build the core product first.  
Do **not** finalize deployment assumptions yet.  
At the end of the build work, explicitly ask:

**“Tell me what environment variables you want and which deployment target you want to use, and I’ll wire the final deployment configuration.”**

That question must come at the end, after the core work is done.

---

## 27) What to Decide Later

Hold these for the final deployment step:
- cloud provider
- object storage vendor
- email provider
- Redis hosting choice
- OCR provider choice if managed
- container/runtime target
- domain/auth callback URLs
- secrets strategy
- monitoring vendor
- CDN / edge setup

Do not hard-bake them now.
