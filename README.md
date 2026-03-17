# project-man

`project-man` is a Go backend for a multi-tenant project management service with a strong security focus.

The project is designed to demonstrate:

- JWT-based authentication
- tenant-aware authorization
- role-based access control
- PostgreSQL Row Level Security (RLS)
- hardened database role configuration
- audited write operations
- repeatable local setup with Docker and Goose migrations
- live endpoint testing, unit tests, and integration tests

This is not a generic CRUD demo. The point of the project is to show how application-layer controls and database-layer controls work together in a secure backend.

## Architecture

The application is structured with a small `cmd/` entrypoint and internal packages for each concern.

Core layout:

- `cmd/server`
  application entrypoint
- `internal/config`
  env-based configuration loading
- `internal/database`
  PostgreSQL connection pool and tenant-scoped transaction helper
- `internal/context`
  request-scoped security context
- `internal/middleware`
  auth, tenant resolution, and RBAC middleware
- `internal/project`
  project handlers, repository logic, and audited mutation flow
- `internal/audit`
  audit log persistence
- `internal/response`
  JSON response helpers
- `db/migrations`
  Goose migrations
- `db/seeds`
  local seed/reset SQL

Project documentation:

- [THREAT_MODEL.md](THREAT_MODEL.md)
- [SECURE_CODE_REVIEW_CHECKLIST.md](SECURE_CODE_REVIEW_CHECKLIST.md)

## Security Model

The project uses multiple layers of enforcement instead of relying on a single check.

### 1. Authentication

`AuthMiddleware` validates a Bearer JWT and rejects unauthenticated requests.

Checks performed:

- expected signing method (`HS256`)
- token expiry
- issuer validation
- audience validation

If valid, a `SecurityContext` is added to the request context with:

- `user_id`
- `email`

Relevant file:

- [auth.go](internal/middleware/auth.go)

### 2. Tenant Resolution

The request must include `X-Tenant-ID`.

The middleware:

- reads `X-Tenant-ID`
- checks the `memberships` table for `(user_id, tenant_id)`
- resolves the role for that tenant
- stores `tenant_id` and `role` in the security context

If the user is not a member of that tenant, the request is rejected with `403`.

Relevant file:

- [tenant.go](internal/middleware/tenant.go)

### 3. Authorization

Authorization is role-based.

Roles:

- `admin`
- `developer`
- `viewer`

Permissions:

- `read_project`
- `create_project`
- `update_project`
- `delete_project`

Relevant file:

- [authz.go](internal/middleware/authz.go)

### 4. Database Tenant Isolation with RLS

The `projects` table is protected by PostgreSQL Row Level Security.

The application sets a transaction-local PostgreSQL setting:

- `app.current_tenant`

Then PostgreSQL enforces:

- only rows with matching `tenant_id` are visible
- only rows with matching `tenant_id` can be inserted or updated

This provides defense in depth:

- application code still scopes queries
- PostgreSQL prevents cross-tenant row access if a query is written incorrectly later

Relevant files:

- [002_enable_rls.sql](db/migrations/002_enable_rls.sql)
- [004_harden_project_role.sql](db/migrations/004_harden_project_role.sql)
- [tenant_tx.go](internal/database/tenant_tx.go)

### 5. Audit Logging

Project mutations are audited.

Audited actions:

- `project.create`
- `project.update`
- `project.delete`

Each audit record stores:

- `actor_user_id`
- `tenant_id`
- `action`
- `resource_type`
- `resource_id`
- `created_at`

Audit records are written in the same transaction as the project mutation.

That means:

- if the mutation fails, no audit record is written
- if the audit insert fails, the mutation is rolled back

Relevant files:

- [003_add_audit_logs.sql](db/migrations/003_add_audit_logs.sql)
- [repository.go](internal/audit/repository.go)
- [service.go](internal/project/service.go)

## API Routes

### Public

- `GET /healthz`

Response:

```json
{"status":"ok"}
```

### Protected

- `GET /secure`
  protected probe route used to verify authenticated + tenant + permission flow

- `POST /projects`
  create a project

- `GET /projects/list`
  list all projects for the active tenant

- `GET /projects/get?id=<project_id>`
  fetch one project in the active tenant

- `PUT /projects/update?id=<project_id>`
  update a project in the active tenant

- `DELETE /projects/delete?id=<project_id>`
  delete a project in the active tenant

Relevant file:

- [main.go](cmd/server/main.go)

## Request Flow

For a protected route, the flow is:

1. request enters the mux
2. `AuthMiddleware` validates JWT and creates the request security context
3. `TenantMiddleware` verifies membership for `X-Tenant-ID`
4. `RequirePermission(...)` checks RBAC
5. handler executes
6. repository/service runs within a tenant-scoped transaction
7. PostgreSQL RLS enforces row visibility and mutation rules
8. audited mutations insert into `audit_logs`

## Database Schema

Current schema:

### `projects`

- `id`
- `tenant_id`
- `name`
- `created_at`

### `memberships`

- `user_id`
- `tenant_id`
- `role`

### `audit_logs`

- `id`
- `actor_user_id`
- `tenant_id`
- `action`
- `resource_type`
- `resource_id`
- `created_at`

Migration files:

- [001_init.sql](db/migrations/001_init.sql)
- [002_enable_rls.sql](db/migrations/002_enable_rls.sql)
- [003_add_audit_logs.sql](db/migrations/003_add_audit_logs.sql)
- [004_harden_project_role.sql](db/migrations/004_harden_project_role.sql)

## Local Development

### Requirements

- Go
- Docker
- Docker Compose

### Environment Variables

The server reads:

- `PORT`
- `DATABASE_URL`
- `JWT_SECRET`
- `JWT_ISSUER`
- `JWT_AUDIENCE`

Defaults used by the local scripts:

```bash
PORT=18080
DATABASE_URL=postgres://project:projectpass@localhost:5432/projectdb
JWT_SECRET=supersecretkey
JWT_ISSUER=project-man
JWT_AUDIENCE=project-man-users
```

### Start Postgres only

```bash
docker compose up -d
```

### Apply migrations and seeds

```bash
./bootstrap_local_db.sh
```

This script:

- runs Goose migrations in `db/migrations`
- applies local seed/reset SQL in `db/seeds`

Relevant file:

- [bootstrap_local_db.sh](bootstrap_local_db.sh)

### Run the server manually

```bash
export GOCACHE=/tmp/go-build
export PORT=18080
export JWT_SECRET=supersecretkey
export JWT_ISSUER=project-man
export JWT_AUDIENCE=project-man-users
export DATABASE_URL=postgres://project:projectpass@localhost:5432/projectdb

go run ./cmd/server/main.go
```

## Testing

### One-command full test cycle

```bash
./run_full_test_cycle.sh
```

This script:

1. starts PostgreSQL with Docker Compose
2. waits for Postgres readiness
3. applies Goose migrations
4. applies seed/reset SQL
5. starts the Go server
6. waits for the health endpoint
7. runs live endpoint tests
8. prints every test case and response in the terminal
9. cleans everything up

Relevant file:

- [run_full_test_cycle.sh](run_full_test_cycle.sh)

### Optional report output

```bash
./run_full_test_cycle.sh /tmp/project-man-report.md
```

### Live endpoint runner only

If the server is already running:

```bash
./run_endpoint_tests.sh
```

Relevant file:

- [run_endpoint_tests.sh](run_endpoint_tests.sh)

### Automated Go tests

```bash
go test ./...
```

Current automated coverage includes:

- JWT auth middleware behavior
- RBAC permission middleware behavior
- project handler validation paths

Relevant test files:

- [auth_test.go](internal/middleware/auth_test.go)
- [authz_test.go](internal/middleware/authz_test.go)
- [handler_test.go](internal/project/handler_test.go)

### Integration tests

```bash
./run_integration_tests.sh
```

This script:

1. starts PostgreSQL with Docker Compose
2. applies Goose migrations and local seeds
3. runs integration tests against the real database
4. cleans up the Docker stack afterward

Current integration coverage includes:

- RLS blocking direct cross-tenant reads
- audit logging for create/update/delete mutations

Relevant files:

- [run_integration_tests.sh](run_integration_tests.sh)
- [integration_test.go](internal/project/integration_test.go)

## CI

The repository includes a GitHub Actions workflow:

- [ci.yml](.github/workflows/ci.yml)

Current CI steps:

- `go test ./...`
- `./run_integration_tests.sh`
- `govulncheck ./...`

## Development JWT Helper

There is a helper program for generating a local development token:

```bash
go run ./generate.go
```

Relevant file:

- [generate.go](generate.go)

This is for local testing only.

## Example Manual Requests

Set a token first:

```bash
TOKEN="$(go run ./generate.go)"
BASE="http://127.0.0.1:18080"
```

Create:

```bash
curl -s -X POST "$BASE/projects" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: tenant-acme" \
  -H "Content-Type: application/json" \
  -d '{"name":"demo-project"}'
```

List:

```bash
curl -s "$BASE/projects/list" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: tenant-acme"
```

## Design Choices

### Why keep both app tenant filters and RLS?

Because they solve different problems:

- app filters make business queries explicit and easy to reason about
- RLS protects against future developer mistakes and missed filters

### Why use `SET LOCAL` tenant context in a transaction?

Because `pgxpool` reuses connections.

A plain session-level setting could leak tenant context between requests.

By using a transaction-local tenant setting, the context is scoped to a single request operation.

### Why harden the PostgreSQL role?

The `project` database role originally had `BYPASSRLS`, which would make RLS ineffective even if policies were defined correctly.

The project now removes that capability through a migration so the integration test proves a real database isolation guarantee instead of a false one.

### Why return `404` for cross-tenant resource access?

This helps avoid leaking resource existence across tenants.

### Why audit only mutations?

This project focuses on sensitive state changes first:

- create
- update
- delete

That gives a clean, high-signal audit trail without overcomplicating the read path.

## Current Strengths

- clear separation between auth, tenant resolution, authz, and business logic
- defense in depth through both app-layer checks and RLS
- verified RLS enforcement after removing `BYPASSRLS` from the DB role
- audited state-changing operations
- reproducible schema management with Goose
- live endpoint testing workflow
- automated unit tests for security-critical middleware and handlers
- integration tests for database-enforced tenant isolation and audit logging

## Known Tradeoffs / Current Limits

- tenant selection is currently client-provided via `X-Tenant-ID`, then server-validated against memberships
- `/secure` is a protected probe route and still returns plain text
- there is no frontend
- CI pipeline and integration tests can still be expanded
- MinIO/file storage is not implemented yet

## Good Next Steps

- add a threat model update if new features like MinIO or file handling are introduced
- extend integration coverage to future DB-backed features
- add a secure code review pass as a repeatable release habit
- expand CI with formatting/lint checks if needed

## Threat Model

A dedicated threat model for this project is available here:

- [THREAT_MODEL.md](THREAT_MODEL.md)

That document covers:

- assets
- trust boundaries
- main threats
- existing controls
- residual risks
- a repeatable threat-modeling method

## Secure Code Review

A repo-specific secure review checklist is available here:

- [SECURE_CODE_REVIEW_CHECKLIST.md](SECURE_CODE_REVIEW_CHECKLIST.md)

Use it to review:

- auth and middleware flow
- tenant and RBAC enforcement
- repository and transaction safety
- migrations and RLS hardening
- audit logging
- tests and local scripts
