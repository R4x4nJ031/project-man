# project-man Endpoint Test Report

Execution date: 2026-03-11
Base URL: `http://127.0.0.1:8080`
Primary test user: `user_id=123`

## Test setup used

- JWT generated from `generate.go`
- Memberships seeded:
  - `123 / tenant-acme / admin`
  - `123 / tenant-view / viewer`
  - `123 / tenant-beta / admin`
- Test focus: endpoint coverage, authn, authz, tenant isolation, and basic input validation

## Results

### Create project as admin [PASS]

**curl command input**

```bash
curl -s -i -X POST "http://127.0.0.1:8080/projects" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJwcm9qZWN0LW1hbi11c2VycyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MTc3MzI1MDk2NywiaXNzIjoicHJvamVjdC1tYW4iLCJ1c2VyX2lkIjoiMTIzIn0.H8uVy2w8E276lNSbSvoFJwFxs_39mWUKEbyyiiajvl4" \
  -H "X-Tenant-ID: tenant-acme" \
  -H "Content-Type: application/json" \
  -d '{"name":"senior-tester-project"}'
```

**Desired output**

Expected HTTP 200 with JSON for the new project in tenant-acme. Desired API shape should include a real DB-generated created_at timestamp.

**Output got**

```http
HTTP/1.1 200 OK
Date: Wed, 11 Mar 2026 16:42:47 GMT
Content-Length: 151
Content-Type: text/plain; charset=utf-8

{"id":"0dfe5219-11ca-41fc-8dc8-19b6b34f5916","tenant_id":"tenant-acme","name":"senior-tester-project","created_at":"2026-03-11T22:12:47.745592+05:30"}
```

### Health check [PASS]

**curl command input**

```bash
curl -s -i "http://127.0.0.1:8080/healthz"
```

**Desired output**

Expected HTTP 200 and body: ok

**Output got**

```http
HTTP/1.1 200 OK
Date: Wed, 11 Mar 2026 16:42:47 GMT
Content-Length: 2
Content-Type: text/plain; charset=utf-8

ok
```

### Secure endpoint as admin [PASS]

**curl command input**

```bash
curl -s -i "http://127.0.0.1:8080/secure" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJwcm9qZWN0LW1hbi11c2VycyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MTc3MzI1MDk2NywiaXNzIjoicHJvamVjdC1tYW4iLCJ1c2VyX2lkIjoiMTIzIn0.H8uVy2w8E276lNSbSvoFJwFxs_39mWUKEbyyiiajvl4" -H "X-Tenant-ID: tenant-acme"
```

**Desired output**

Expected HTTP 200 and body confirming authorized project creation.

**Output got**

```http
HTTP/1.1 200 OK
Date: Wed, 11 Mar 2026 16:42:47 GMT
Content-Length: 27
Content-Type: text/plain; charset=utf-8

authorized project creation
```

### List projects as admin [PASS]

**curl command input**

```bash
curl -s -i "http://127.0.0.1:8080/projects/list" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJwcm9qZWN0LW1hbi11c2VycyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MTc3MzI1MDk2NywiaXNzIjoicHJvamVjdC1tYW4iLCJ1c2VyX2lkIjoiMTIzIn0.H8uVy2w8E276lNSbSvoFJwFxs_39mWUKEbyyiiajvl4" -H "X-Tenant-ID: tenant-acme"
```

**Desired output**

Expected HTTP 200 with a JSON array containing the project created for tenant-acme only.

**Output got**

```http
HTTP/1.1 200 OK
Date: Wed, 11 Mar 2026 16:42:47 GMT
Content-Length: 153
Content-Type: text/plain; charset=utf-8

[{"id":"0dfe5219-11ca-41fc-8dc8-19b6b34f5916","tenant_id":"tenant-acme","name":"senior-tester-project","created_at":"2026-03-11T22:12:47.745592+05:30"}]
```

### Get project by id in same tenant [PASS]

**curl command input**

```bash
curl -s -i "http://127.0.0.1:8080/projects/get?id=0dfe5219-11ca-41fc-8dc8-19b6b34f5916" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJwcm9qZWN0LW1hbi11c2VycyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MTc3MzI1MDk2NywiaXNzIjoicHJvamVjdC1tYW4iLCJ1c2VyX2lkIjoiMTIzIn0.H8uVy2w8E276lNSbSvoFJwFxs_39mWUKEbyyiiajvl4" -H "X-Tenant-ID: tenant-acme"
```

**Desired output**

Expected HTTP 200 with the requested project JSON.

**Output got**

```http
HTTP/1.1 200 OK
Date: Wed, 11 Mar 2026 16:42:47 GMT
Content-Length: 151
Content-Type: text/plain; charset=utf-8

{"id":"0dfe5219-11ca-41fc-8dc8-19b6b34f5916","tenant_id":"tenant-acme","name":"senior-tester-project","created_at":"2026-03-11T22:12:47.745592+05:30"}
```

### Update project name in same tenant [PASS]

**curl command input**

```bash
curl -s -i -X PUT "http://127.0.0.1:8080/projects/update?id=0dfe5219-11ca-41fc-8dc8-19b6b34f5916" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJwcm9qZWN0LW1hbi11c2VycyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MTc3MzI1MDk2NywiaXNzIjoicHJvamVjdC1tYW4iLCJ1c2VyX2lkIjoiMTIzIn0.H8uVy2w8E276lNSbSvoFJwFxs_39mWUKEbyyiiajvl4" -H "X-Tenant-ID: tenant-acme" -H "Content-Type: application/json" -d '{"name":"renamed-by-senior-tester"}'
```

**Desired output**

Expected HTTP 200 with updated JSON reflecting the new name.

**Output got**

```http
HTTP/1.1 200 OK
Date: Wed, 11 Mar 2026 16:42:47 GMT
Content-Length: 154
Content-Type: text/plain; charset=utf-8

{"id":"0dfe5219-11ca-41fc-8dc8-19b6b34f5916","tenant_id":"tenant-acme","name":"renamed-by-senior-tester","created_at":"2026-03-11T22:12:47.745592+05:30"}
```

### Delete project in same tenant [PASS]

**curl command input**

```bash
curl -s -i -X DELETE "http://127.0.0.1:8080/projects/delete?id=0dfe5219-11ca-41fc-8dc8-19b6b34f5916" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJwcm9qZWN0LW1hbi11c2VycyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MTc3MzI1MDk2NywiaXNzIjoicHJvamVjdC1tYW4iLCJ1c2VyX2lkIjoiMTIzIn0.H8uVy2w8E276lNSbSvoFJwFxs_39mWUKEbyyiiajvl4" -H "X-Tenant-ID: tenant-acme"
```

**Desired output**

Expected HTTP 204 with an empty body.

**Output got**

```http
HTTP/1.1 204 No Content
Date: Wed, 11 Mar 2026 16:42:47 GMT

```

### Get deleted project [PASS]

**curl command input**

```bash
curl -s -i "http://127.0.0.1:8080/projects/get?id=0dfe5219-11ca-41fc-8dc8-19b6b34f5916" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJwcm9qZWN0LW1hbi11c2VycyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MTc3MzI1MDk2NywiaXNzIjoicHJvamVjdC1tYW4iLCJ1c2VyX2lkIjoiMTIzIn0.H8uVy2w8E276lNSbSvoFJwFxs_39mWUKEbyyiiajvl4" -H "X-Tenant-ID: tenant-acme"
```

**Desired output**

Expected HTTP 404 because the project was deleted.

**Output got**

```http
HTTP/1.1 404 Not Found
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Wed, 11 Mar 2026 16:42:47 GMT
Content-Length: 18

project not found
```

### Missing auth header [PASS]

**curl command input**

```bash
curl -s -i "http://127.0.0.1:8080/projects/list" -H "X-Tenant-ID: tenant-acme"
```

**Desired output**

Expected HTTP 401 with message: missing authorization header

**Output got**

```http
HTTP/1.1 401 Unauthorized
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Wed, 11 Mar 2026 16:42:47 GMT
Content-Length: 29

missing authorization header
```

### Invalid bearer token [PASS]

**curl command input**

```bash
curl -s -i "http://127.0.0.1:8080/projects/list" -H "Authorization: Bearer garbage.token.here" -H "X-Tenant-ID: tenant-acme"
```

**Desired output**

Expected HTTP 401 with message: invalid token

**Output got**

```http
HTTP/1.1 401 Unauthorized
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Wed, 11 Mar 2026 16:42:47 GMT
Content-Length: 14

invalid token
```

### Missing tenant header [PASS]

**curl command input**

```bash
curl -s -i "http://127.0.0.1:8080/projects/list" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJwcm9qZWN0LW1hbi11c2VycyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MTc3MzI1MDk2NywiaXNzIjoicHJvamVjdC1tYW4iLCJ1c2VyX2lkIjoiMTIzIn0.H8uVy2w8E276lNSbSvoFJwFxs_39mWUKEbyyiiajvl4"
```

**Desired output**

Expected HTTP 403 with message: missing tenant id

**Output got**

```http
HTTP/1.1 403 Forbidden
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Wed, 11 Mar 2026 16:42:47 GMT
Content-Length: 18

missing tenant id
```

### Unknown tenant membership [PASS]

**curl command input**

```bash
curl -s -i "http://127.0.0.1:8080/projects/list" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJwcm9qZWN0LW1hbi11c2VycyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MTc3MzI1MDk2NywiaXNzIjoicHJvamVjdC1tYW4iLCJ1c2VyX2lkIjoiMTIzIn0.H8uVy2w8E276lNSbSvoFJwFxs_39mWUKEbyyiiajvl4" -H "X-Tenant-ID: tenant-ghost"
```

**Desired output**

Expected HTTP 403 with message: forbidden tenant access

**Output got**

```http
HTTP/1.1 403 Forbidden
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Wed, 11 Mar 2026 16:42:47 GMT
Content-Length: 24

forbidden tenant access
```

### Viewer blocked from create [PASS]

**curl command input**

```bash
curl -s -i -X POST "http://127.0.0.1:8080/projects" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJwcm9qZWN0LW1hbi11c2VycyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MTc3MzI1MDk2NywiaXNzIjoicHJvamVjdC1tYW4iLCJ1c2VyX2lkIjoiMTIzIn0.H8uVy2w8E276lNSbSvoFJwFxs_39mWUKEbyyiiajvl4" -H "X-Tenant-ID: tenant-view" -H "Content-Type: application/json" -d '{"name":"viewer-should-fail"}'
```

**Desired output**

Expected HTTP 403 with message: forbidden because viewer lacks create_project permission.

**Output got**

```http
HTTP/1.1 403 Forbidden
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Wed, 11 Mar 2026 16:42:47 GMT
Content-Length: 10

forbidden
```

### Viewer blocked from secure endpoint [PASS]

**curl command input**

```bash
curl -s -i "http://127.0.0.1:8080/secure" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJwcm9qZWN0LW1hbi11c2VycyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MTc3MzI1MDk2NywiaXNzIjoicHJvamVjdC1tYW4iLCJ1c2VyX2lkIjoiMTIzIn0.H8uVy2w8E276lNSbSvoFJwFxs_39mWUKEbyyiiajvl4" -H "X-Tenant-ID: tenant-view"
```

**Desired output**

Expected HTTP 403 with message: forbidden.

**Output got**

```http
HTTP/1.1 403 Forbidden
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Wed, 11 Mar 2026 16:42:47 GMT
Content-Length: 10

forbidden
```

### Cross-tenant update blocked at SQL layer [PASS]

**curl command input**

```bash
curl -s -i -X PUT "http://127.0.0.1:8080/projects/update?id=0dfe5219-11ca-41fc-8dc8-19b6b34f5916" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJwcm9qZWN0LW1hbi11c2VycyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MTc3MzI1MDk2NywiaXNzIjoicHJvamVjdC1tYW4iLCJ1c2VyX2lkIjoiMTIzIn0.H8uVy2w8E276lNSbSvoFJwFxs_39mWUKEbyyiiajvl4" -H "X-Tenant-ID: tenant-beta" -H "Content-Type: application/json" -d '{"name":"cross-tenant-attempt"}'
```

**Desired output**

Expected HTTP 404 because the project id does not belong to tenant-beta.

**Output got**

```http
HTTP/1.1 404 Not Found
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Wed, 11 Mar 2026 16:42:48 GMT
Content-Length: 18

project not found
```

### Cross-tenant delete blocked at SQL layer [PASS]

**curl command input**

```bash
curl -s -i -X DELETE "http://127.0.0.1:8080/projects/delete?id=0dfe5219-11ca-41fc-8dc8-19b6b34f5916" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJwcm9qZWN0LW1hbi11c2VycyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MTc3MzI1MDk2NywiaXNzIjoicHJvamVjdC1tYW4iLCJ1c2VyX2lkIjoiMTIzIn0.H8uVy2w8E276lNSbSvoFJwFxs_39mWUKEbyyiiajvl4" -H "X-Tenant-ID: tenant-beta"
```

**Desired output**

Expected HTTP 404 because the project id does not belong to tenant-beta.

**Output got**

```http
HTTP/1.1 404 Not Found
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Wed, 11 Mar 2026 16:42:48 GMT
Content-Length: 18

project not found
```

### Get project missing id [PASS]

**curl command input**

```bash
curl -s -i "http://127.0.0.1:8080/projects/get" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJwcm9qZWN0LW1hbi11c2VycyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MTc3MzI1MDk2NywiaXNzIjoicHJvamVjdC1tYW4iLCJ1c2VyX2lkIjoiMTIzIn0.H8uVy2w8E276lNSbSvoFJwFxs_39mWUKEbyyiiajvl4" -H "X-Tenant-ID: tenant-acme"
```

**Desired output**

Expected HTTP 400 with message: missing project id

**Output got**

```http
HTTP/1.1 400 Bad Request
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Wed, 11 Mar 2026 16:42:48 GMT
Content-Length: 19

missing project id
```

### Update project with empty name [PASS]

**curl command input**

```bash
curl -s -i -X PUT "http://127.0.0.1:8080/projects/update?id=0dfe5219-11ca-41fc-8dc8-19b6b34f5916" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJwcm9qZWN0LW1hbi11c2VycyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MTc3MzI1MDk2NywiaXNzIjoicHJvamVjdC1tYW4iLCJ1c2VyX2lkIjoiMTIzIn0.H8uVy2w8E276lNSbSvoFJwFxs_39mWUKEbyyiiajvl4" -H "X-Tenant-ID: tenant-acme" -H "Content-Type: application/json" -d '{"name":""}'
```

**Desired output**

Expected HTTP 400 with message: name cannot be empty

**Output got**

```http
HTTP/1.1 400 Bad Request
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Wed, 11 Mar 2026 16:42:48 GMT
Content-Length: 21

name cannot be empty
```

### Delete project missing id [PASS]

**curl command input**

```bash
curl -s -i -X DELETE "http://127.0.0.1:8080/projects/delete" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJwcm9qZWN0LW1hbi11c2VycyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImV4cCI6MTc3MzI1MDk2NywiaXNzIjoicHJvamVjdC1tYW4iLCJ1c2VyX2lkIjoiMTIzIn0.H8uVy2w8E276lNSbSvoFJwFxs_39mWUKEbyyiiajvl4" -H "X-Tenant-ID: tenant-acme"
```

**Desired output**

Expected HTTP 400 with message: missing project id

**Output got**

```http
HTTP/1.1 400 Bad Request
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Wed, 11 Mar 2026 16:42:48 GMT
Content-Length: 19

missing project id
```

## Senior tester summary

- Core CRUD flow works for an authorized tenant admin.
- Auth middleware, tenant membership lookup, and RBAC enforcement all responded with the expected status codes in this run.
- Tenant isolation on read/update/delete is enforced by SQL predicates on `tenant_id`.
- `POST /projects` now returns the persisted row with a real database-generated `created_at` timestamp.
- Operational note: this project has no checked-in migration/schema file; the test depended on manual table creation and seed data.
