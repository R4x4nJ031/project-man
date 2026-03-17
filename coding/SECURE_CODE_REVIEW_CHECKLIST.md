# Secure Code Review Checklist

This checklist is specific to `project-man`.

Use it when reviewing this repository for security issues, regressions, or risky design changes.

The goal is not to ask every possible question. The goal is to ask the highest-value questions in the right places.

## How To Use This Checklist

Read the project in this order:

1. [README.md](/home/raxan/Downloads/coding/README.md)
2. [THREAT_MODEL.md](/home/raxan/Downloads/coding/THREAT_MODEL.md)
3. [main.go](/home/raxan/Downloads/coding/cmd/server/main.go)
4. `internal/middleware`
5. `internal/project`
6. `internal/database`
7. `db/migrations`
8. tests and scripts

That order matters because secure review is easier when you understand:

- what the system does
- where trust boundaries are
- how requests flow
- where the DB safety nets live

## 1. Entry Point Review

Files:

- [main.go](/home/raxan/Downloads/coding/cmd/server/main.go)

Questions:

- Are all protected routes actually wrapped in auth middleware?
- Is middleware order correct?
- Can any business handler be reached without auth, tenant resolution, and permission checks?
- Are there any routes that accidentally bypass tenant checks?
- Are any sensitive routes left public?

Expected answer in this project:

- `/healthz` is public
- `/secure` and `/projects*` are protected
- auth runs before tenant resolution
- tenant resolution runs before RBAC

## 2. Authentication Review

Files:

- [auth.go](/home/raxan/Downloads/coding/internal/middleware/auth.go)

Questions:

- Is the JWT signature verified?
- Is the signing algorithm validated explicitly?
- Are expiry, issuer, and audience checked?
- Are invalid tokens rejected with the right status?
- Is user identity extracted only after successful validation?
- Are token details or secrets accidentally logged?

Things to watch for:

- accepting `alg=none`
- skipping issuer or audience validation
- trusting claims before `token.Valid`
- weak secret handling

## 3. Tenant Resolution Review

Files:

- [tenant.go](/home/raxan/Downloads/coding/internal/middleware/tenant.go)
- [security_context.go](/home/raxan/Downloads/coding/internal/context/security_context.go)

Questions:

- Is `X-Tenant-ID` treated as untrusted input first?
- Is tenant membership checked in the DB before trust is granted?
- Is role resolved server-side instead of trusted from the client?
- Does the middleware reject unknown tenant membership with `403`?
- Does the resolved tenant end up in request context safely?

Things to watch for:

- trusting client-supplied tenant values directly
- skipping membership lookup
- deriving role from the token or client without verification

## 4. Authorization Review

Files:

- [authz.go](/home/raxan/Downloads/coding/internal/middleware/authz.go)

Questions:

- Is permission checking deny-by-default?
- Are unknown roles blocked?
- Are role-to-permission mappings sensible?
- Are mutation routes protected by mutation permissions?

Things to watch for:

- permissive fallbacks
- missing checks on a new route
- role map drift after feature expansion

## 5. Handler Review

Files:

- [handler.go](/home/raxan/Downloads/coding/internal/project/handler.go)

Questions:

- Do handlers require security context before use?
- Are query params and JSON bodies validated?
- Are input errors handled without leaking internals?
- Are handlers delegating DB work rather than embedding SQL?
- Do handlers avoid making security decisions from client-controlled fields alone?

Things to watch for:

- missing ID validation
- missing body validation
- returning raw DB errors to clients
- mutation logic skipping service/repository paths

## 6. Business Logic Review

Files:

- [service.go](/home/raxan/Downloads/coding/internal/project/service.go)

Questions:

- Are project mutations and audit writes performed in the same transaction?
- If the audit write fails, does the mutation roll back?
- Are actor identity and tenant passed in from trusted context?
- Is business logic isolated from transport details?

Things to watch for:

- “best effort” audit logging
- mutation succeeds without audit
- business logic depending on client input instead of trusted context

## 7. Repository Review

Files:

- [repository.go](/home/raxan/Downloads/coding/internal/project/repository.go)
- [tenant_tx.go](/home/raxan/Downloads/coding/internal/database/tenant_tx.go)

Questions:

- Do project operations run inside tenant-scoped transactions?
- Is the tenant context set with transaction-local scope?
- Are queries still tenant-aware even with RLS present?
- Are read/update/delete queries scoped to the expected tenant?
- Are no-row cases handled safely?

Things to watch for:

- raw access to `database.DB` without tenant transaction wrapper
- session-level tenant context leakage across pooled connections
- queries that assume RLS is enough and become unclear

## 8. Database Review

Files:

- [001_init.sql](/home/raxan/Downloads/coding/db/migrations/001_init.sql)
- [002_enable_rls.sql](/home/raxan/Downloads/coding/db/migrations/002_enable_rls.sql)
- [003_add_audit_logs.sql](/home/raxan/Downloads/coding/db/migrations/003_add_audit_logs.sql)
- [004_harden_project_role.sql](/home/raxan/Downloads/coding/db/migrations/004_harden_project_role.sql)

Questions:

- Is schema evolution tracked through migrations?
- Is RLS enabled and forced on tenant-owned tables?
- Are RLS policies aligned with tenant context?
- Is the DB role hardened so RLS cannot be bypassed?
- Are audit tables append-only by design?

Things to watch for:

- `BYPASSRLS`
- unprotected new tenant-owned tables
- migrations that weaken existing security guarantees

## 9. Auditability Review

Files:

- [repository.go](/home/raxan/Downloads/coding/internal/audit/repository.go)
- [003_add_audit_logs.sql](/home/raxan/Downloads/coding/db/migrations/003_add_audit_logs.sql)

Questions:

- Are all sensitive mutations audited?
- Does each entry contain actor, tenant, action, resource type, and resource id?
- Are audit entries timestamped automatically?
- Can audit rows be silently overwritten or deleted by application code?

Things to watch for:

- missing mutation paths
- incomplete metadata
- audit writes outside the main transaction

## 10. Configuration Review

Files:

- [config.go](/home/raxan/Downloads/coding/internal/config/config.go)
- [generate.go](/home/raxan/Downloads/coding/generate.go)

Questions:

- Are secrets loaded from env rather than hardcoded in runtime code?
- Are required env vars enforced?
- Is the token generator clearly dev-only?
- Are defaults reasonable for development but not misleading for production?

Things to watch for:

- hardcoded secrets in normal code paths
- production-looking defaults that are insecure

## 11. Test Review

Files:

- [auth_test.go](/home/raxan/Downloads/coding/internal/middleware/auth_test.go)
- [authz_test.go](/home/raxan/Downloads/coding/internal/middleware/authz_test.go)
- [handler_test.go](/home/raxan/Downloads/coding/internal/project/handler_test.go)
- [integration_test.go](/home/raxan/Downloads/coding/internal/project/integration_test.go)

Questions:

- Do tests cover both allow and deny paths?
- Are the highest-risk controls tested?
- Is RLS tested against a real DB?
- Is audit logging tested against a real DB?
- Do tests prove actual behavior, not just happy-path assumptions?

Things to watch for:

- tests only for success cases
- fake tests that never exercise the security boundary
- no regression tests for previously found issues

## 12. Scripts / Operational Review

Files:

- [bootstrap_local_db.sh](/home/raxan/Downloads/coding/bootstrap_local_db.sh)
- [run_full_test_cycle.sh](/home/raxan/Downloads/coding/run_full_test_cycle.sh)
- [run_endpoint_tests.sh](/home/raxan/Downloads/coding/run_endpoint_tests.sh)
- [run_integration_tests.sh](/home/raxan/Downloads/coding/run_integration_tests.sh)

Questions:

- Do local scripts reflect the actual supported setup path?
- Do scripts apply migrations before running tests?
- Are cleanup behaviors sane?
- Are dev-only secrets and assumptions clearly local only?

Things to watch for:

- stale scripts
- setup paths that skip migrations
- tests accidentally hitting old local servers

## 13. Highest-Risk Review Questions

If you only have a short time, ask these first:

1. Can any route be reached without auth?
2. Can any route be reached without tenant membership verification?
3. Can a viewer mutate project state?
4. Can cross-tenant data be accessed if a query forgets tenant filtering?
5. Does the DB role bypass RLS?
6. Are create/update/delete actions audited transactionally?
7. Are secrets hardcoded or leaked?

If all 7 are answered well, the review is already meaningful.

## 14. What “Good” Looks Like In This Repo

A strong review outcome for this project should conclude:

- auth is enforced consistently
- tenant membership is validated before trust
- RBAC is deny-by-default
- project mutations are auditable
- RLS is enabled and not bypassable by the DB role
- integration tests prove critical security guarantees

If any of those fail, that is a serious finding.
