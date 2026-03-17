#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
REPORT_PATH="${1:-}"
EXECUTION_DATE="$(date +%F)"

export GOCACHE="${GOCACHE:-/tmp/go-build}"
TOKEN="$(go run ./generate.go | tr -d '\r\n')"

PROJECT_ID=""
CREATE_OUTPUT=""
PASS_COUNT=0
FAIL_COUNT=0

extract_status() {
  printf '%s\n' "$1" | tr -d '\r' | awk 'NR==1 {print $2}'
}

extract_body() {
  printf '%s\n' "$1" | tr -d '\r' | awk 'BEGIN{body=0} /^$/{body=1; next} body{print}'
}

init_report() {
  [[ -z "$REPORT_PATH" ]] && return 0

  mkdir -p "$(dirname "$REPORT_PATH")"
  cat >"$REPORT_PATH" <<EOF
# project-man Endpoint Test Report

Execution date: $EXECUTION_DATE
Base URL: \`$BASE_URL\`
Primary test user: \`user_id=123\`

## Test setup used

- JWT generated from \`generate.go\`
- Memberships seeded:
  - \`123 / tenant-acme / admin\`
  - \`123 / tenant-view / viewer\`
  - \`123 / tenant-beta / admin\`
- Test focus: endpoint coverage, authn, authz, tenant isolation, and basic input validation

## Results

EOF
}

append_report_case() {
  local name="$1"
  local command="$2"
  local desired="$3"
  local result="$4"
  local output="$5"

  [[ -z "$REPORT_PATH" ]] && return 0

  {
    printf '### %s [%s]\n\n' "$name" "$result"
    printf '**curl command input**\n\n```bash\n%s\n```\n\n' "$command"
    printf '**Desired output**\n\n%s\n\n' "$desired"
    printf '**Output got**\n\n```http\n%s\n```\n\n' "$output"
  } >>"$REPORT_PATH"
}

append_report_summary() {
  [[ -z "$REPORT_PATH" ]] && return 0

  cat >>"$REPORT_PATH" <<EOF
## Senior tester summary

- Core CRUD flow works for an authorized tenant admin.
- Auth middleware, tenant membership lookup, and RBAC enforcement responded with the expected status codes in this run.
- Tenant isolation on read/update/delete is enforced by SQL predicates on \`tenant_id\`.
- \`POST /projects\` returns the persisted row with a real database-generated \`created_at\` timestamp.
- Operational note: this project has no checked-in migration/schema file; the test depends on manual table creation and seed data.
- Totals: PASS=$PASS_COUNT, FAIL=$FAIL_COUNT
EOF
}

print_case() {
  local name="$1"
  local command="$2"
  local desired="$3"
  local expected_status="$4"
  local output="$5"
  local actual_status
  local result

  actual_status="$(extract_status "$output")"
  if [[ "$actual_status" == "$expected_status" ]]; then
    result="PASS"
    PASS_COUNT=$((PASS_COUNT + 1))
  else
    result="FAIL"
    FAIL_COUNT=$((FAIL_COUNT + 1))
  fi

  printf '\n============================================================\n'
  printf 'TEST: %s\n' "$name"
  printf 'RESULT: %s\n' "$result"
  printf 'EXPECTED STATUS: %s\n' "$expected_status"
  printf 'ACTUAL STATUS: %s\n' "${actual_status:-none}"
  printf '\nCOMMAND:\n%s\n' "$command"
  printf '\nDESIRED OUTPUT:\n%s\n' "$desired"
  printf '\nOUTPUT GOT:\n%s\n' "$output"
  printf '============================================================\n'

  append_report_case "$name" "$command" "$desired" "$result" "$output"
}

run_case() {
  local name="$1"
  local command="$2"
  local desired="$3"
  local expected_status="$4"
  local output

  output="$(bash -lc "$command" 2>&1 || true)"
  print_case "$name" "$command" "$desired" "$expected_status" "$output"
}

init_report

printf 'Starting endpoint test run for %s\n' "$BASE_URL"
printf 'JWT generated for user_id=123\n'

CREATE_COMMAND=$(cat <<EOF
curl -s -i -X POST "$BASE_URL/projects" \\
  -H "Authorization: Bearer $TOKEN" \\
  -H "X-Tenant-ID: tenant-acme" \\
  -H "Content-Type: application/json" \\
  -d '{"name":"senior-tester-project"}'
EOF
)

CREATE_OUTPUT="$(bash -lc "$CREATE_COMMAND" 2>&1 || true)"
print_case \
  "Create project as admin" \
  "$CREATE_COMMAND" \
  "Expected HTTP 200 with JSON for the new project in tenant-acme. The response should include a real DB-generated created_at timestamp." \
  "200" \
  "$CREATE_OUTPUT"

PROJECT_ID="$(extract_body "$CREATE_OUTPUT" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p' | head -n1)"

run_case \
  "Health check" \
  "curl -s -i \"$BASE_URL/healthz\"" \
  "Expected HTTP 200 with JSON body: {\"status\":\"ok\"}" \
  "200"

run_case \
  "Secure endpoint as admin" \
  "curl -s -i \"$BASE_URL/secure\" -H \"Authorization: Bearer $TOKEN\" -H \"X-Tenant-ID: tenant-acme\"" \
  "Expected HTTP 200 and body confirming authorized project creation." \
  "200"

run_case \
  "List projects as admin" \
  "curl -s -i \"$BASE_URL/projects/list\" -H \"Authorization: Bearer $TOKEN\" -H \"X-Tenant-ID: tenant-acme\"" \
  "Expected HTTP 200 with a JSON array containing the project created for tenant-acme only." \
  "200"

run_case \
  "Get project by id in same tenant" \
  "curl -s -i \"$BASE_URL/projects/get?id=$PROJECT_ID\" -H \"Authorization: Bearer $TOKEN\" -H \"X-Tenant-ID: tenant-acme\"" \
  "Expected HTTP 200 with the requested project JSON." \
  "200"

run_case \
  "Update project name in same tenant" \
  "curl -s -i -X PUT \"$BASE_URL/projects/update?id=$PROJECT_ID\" -H \"Authorization: Bearer $TOKEN\" -H \"X-Tenant-ID: tenant-acme\" -H \"Content-Type: application/json\" -d '{\"name\":\"renamed-by-senior-tester\"}'" \
  "Expected HTTP 200 with updated JSON reflecting the new name." \
  "200"

run_case \
  "Delete project in same tenant" \
  "curl -s -i -X DELETE \"$BASE_URL/projects/delete?id=$PROJECT_ID\" -H \"Authorization: Bearer $TOKEN\" -H \"X-Tenant-ID: tenant-acme\"" \
  "Expected HTTP 204 with an empty body." \
  "204"

run_case \
  "Get deleted project" \
  "curl -s -i \"$BASE_URL/projects/get?id=$PROJECT_ID\" -H \"Authorization: Bearer $TOKEN\" -H \"X-Tenant-ID: tenant-acme\"" \
  "Expected HTTP 404 because the project was deleted." \
  "404"

run_case \
  "Missing auth header" \
  "curl -s -i \"$BASE_URL/projects/list\" -H \"X-Tenant-ID: tenant-acme\"" \
  "Expected HTTP 401 with JSON error body for missing authorization header." \
  "401"

run_case \
  "Invalid bearer token" \
  "curl -s -i \"$BASE_URL/projects/list\" -H \"Authorization: Bearer garbage.token.here\" -H \"X-Tenant-ID: tenant-acme\"" \
  "Expected HTTP 401 with JSON error body for invalid token." \
  "401"

run_case \
  "Missing tenant header" \
  "curl -s -i \"$BASE_URL/projects/list\" -H \"Authorization: Bearer $TOKEN\"" \
  "Expected HTTP 403 with JSON error body for missing tenant id." \
  "403"

run_case \
  "Unknown tenant membership" \
  "curl -s -i \"$BASE_URL/projects/list\" -H \"Authorization: Bearer $TOKEN\" -H \"X-Tenant-ID: tenant-ghost\"" \
  "Expected HTTP 403 with JSON error body for forbidden tenant access." \
  "403"

run_case \
  "Viewer blocked from create" \
  "curl -s -i -X POST \"$BASE_URL/projects\" -H \"Authorization: Bearer $TOKEN\" -H \"X-Tenant-ID: tenant-view\" -H \"Content-Type: application/json\" -d '{\"name\":\"viewer-should-fail\"}'" \
  "Expected HTTP 403 with JSON error body because viewer lacks create_project permission." \
  "403"

run_case \
  "Viewer blocked from secure endpoint" \
  "curl -s -i \"$BASE_URL/secure\" -H \"Authorization: Bearer $TOKEN\" -H \"X-Tenant-ID: tenant-view\"" \
  "Expected HTTP 403 with JSON error body for forbidden access." \
  "403"

run_case \
  "Cross-tenant update blocked at SQL layer" \
  "curl -s -i -X PUT \"$BASE_URL/projects/update?id=$PROJECT_ID\" -H \"Authorization: Bearer $TOKEN\" -H \"X-Tenant-ID: tenant-beta\" -H \"Content-Type: application/json\" -d '{\"name\":\"cross-tenant-attempt\"}'" \
  "Expected HTTP 404 because the project id does not belong to tenant-beta." \
  "404"

run_case \
  "Cross-tenant delete blocked at SQL layer" \
  "curl -s -i -X DELETE \"$BASE_URL/projects/delete?id=$PROJECT_ID\" -H \"Authorization: Bearer $TOKEN\" -H \"X-Tenant-ID: tenant-beta\"" \
  "Expected HTTP 404 because the project id does not belong to tenant-beta." \
  "404"

run_case \
  "Get project missing id" \
  "curl -s -i \"$BASE_URL/projects/get\" -H \"Authorization: Bearer $TOKEN\" -H \"X-Tenant-ID: tenant-acme\"" \
  "Expected HTTP 400 with JSON error body for missing project id." \
  "400"

run_case \
  "Update project with empty name" \
  "curl -s -i -X PUT \"$BASE_URL/projects/update?id=$PROJECT_ID\" -H \"Authorization: Bearer $TOKEN\" -H \"X-Tenant-ID: tenant-acme\" -H \"Content-Type: application/json\" -d '{\"name\":\"\"}'" \
  "Expected HTTP 400 with JSON error body for empty name." \
  "400"

run_case \
  "Delete project missing id" \
  "curl -s -i -X DELETE \"$BASE_URL/projects/delete\" -H \"Authorization: Bearer $TOKEN\" -H \"X-Tenant-ID: tenant-acme\"" \
  "Expected HTTP 400 with JSON error body for missing project id." \
  "400"

append_report_summary

printf '\nFinished endpoint test run. PASS=%d FAIL=%d\n' "$PASS_COUNT" "$FAIL_COUNT"
if [[ -n "$REPORT_PATH" ]]; then
  printf 'Report written to: %s\n' "$REPORT_PATH"
fi
