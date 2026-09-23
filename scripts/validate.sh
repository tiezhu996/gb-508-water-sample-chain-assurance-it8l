#!/usr/bin/env sh
set -eu
project_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$project_root"
set -a
if [ -f .env ]; then . ./.env; else . ./.env.example; fi
set +a
(command -v jq >/dev/null 2>&1) || { echo "jq is required for API validation" >&2; exit 1; }
(cd backend && go test ./... && go build ./...)
(cd frontend && npm install --no-audit --no-fund && npm run build)
docker compose config --quiet
docker compose up -d --build
cleanup() { docker compose down -v --remove-orphans; }
if [ "${KEEP_RUNNING:-0}" = "1" ]; then
  trap cleanup INT TERM
else
  trap cleanup EXIT INT TERM
fi
i=0
until curl -fsS "http://127.0.0.1:${BACKEND_PORT:-19508}/healthz" >/dev/null; do
  i=$((i+1)); [ "$i" -lt 60 ] || { docker compose logs; exit 1; }; sleep 2
done
curl -fsS "http://127.0.0.1:${FRONTEND_PORT:-18508}/" >/dev/null
token=$(curl -fsS -X POST "http://127.0.0.1:${BACKEND_PORT:-19508}/api/auth/login" -H 'Content-Type: application/json' -d '{"username":"admin","password":"Admin123!"}' | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
[ -n "$token" ]
operator_token=$(curl -fsS -X POST "http://127.0.0.1:${BACKEND_PORT:-19508}/api/auth/login" -H 'Content-Type: application/json' -d '{"username":"operator","password":"Admin123!"}' | jq -er '.data.token')
reviewer_token=$(curl -fsS -X POST "http://127.0.0.1:${BACKEND_PORT:-19508}/api/auth/login" -H 'Content-Type: application/json' -d '{"username":"reviewer","password":"Admin123!"}' | jq -er '.data.token')
curl -fsS "http://127.0.0.1:${BACKEND_PORT:-19508}/api/overview" -H "Authorization: Bearer $token" >/dev/null
curl -fsS "http://127.0.0.1:${BACKEND_PORT}/api/session" -H "Authorization: Bearer $token" | jq -e '.data.role == "admin" and (.data.requestId | length > 0)' >/dev/null
curl -fsS "http://127.0.0.1:${BACKEND_PORT}/api/runtime" -H "Authorization: Bearer $token" | jq -e '.data.appName and .data.databaseDriver and (.data.requestLimit > 0)' >/dev/null
paths=$(sed -n "s/.*path: '\\([^']*\\)'.*/\\1/p" frontend/src/types/status.ts)
for path in $paths; do
  curl -fsS "http://127.0.0.1:${BACKEND_PORT}/api/$path?page=1&pageSize=20" -H "Authorization: Bearer $token" | jq -e '.data | type == "array"' >/dev/null
done
entity_config=$(sed -n "s/.*path: '\\([^']*\\)'.*statuses: \\['\\([^']*\\)', '\\([^']*\\)'.*/\\1|\\2|\\3/p" frontend/src/types/status.ts | head -n 1)
resource=$(printf '%s' "$entity_config" | cut -d '|' -f 1)
initial_status=$(printf '%s' "$entity_config" | cut -d '|' -f 2)
next_status=$(printf '%s' "$entity_config" | cut -d '|' -f 3)
now=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
code="SMOKE-$(date +%s)"
payload=$(printf '{"code":"%s","name":"Runtime smoke record","description":"Automated Compose workflow validation","facility":"Validation Lab","owner":"admin","category":"smoke","riskLevel":"low","metricValue":1,"metricUnit":"unit","effectiveAt":"%s","evidence":"scripts/validate.sh","relatedCode":"SMOKE"}' "$code" "$now")
created=$(curl -fsS -X POST "http://127.0.0.1:${BACKEND_PORT}/api/$resource" -H "Authorization: Bearer $token" -H 'Content-Type: application/json' -d "$payload")
id=$(printf '%s' "$created" | jq -er '.data.id')
version=$(printf '%s' "$created" | jq -er '.data.version')
printf '%s' "$created" | jq -e --arg status "$initial_status" '.data.status == $status' >/dev/null
transition=$(printf '{"status":"%s","expectedVersion":%s,"reason":"automated runtime validation"}' "$next_status" "$version")
curl -fsS -X POST "http://127.0.0.1:${BACKEND_PORT}/api/$resource/$id/transition" -H "Authorization: Bearer $token" -H 'Content-Type: application/json' -d "$transition" | jq -e --arg status "$next_status" '.data.status == $status' >/dev/null
review_code="REVIEW-SMOKE-$(date +%s)-$$"
review_payload=$(printf '{"code":"%s","name":"Dual reviewer validation","description":"Validates the required peer-review workflow","facility":"Validation Lab","owner":"operator","category":"smoke","riskLevel":"low","metricValue":1,"metricUnit":"unit","effectiveAt":"%s","evidence":"scripts/validate.sh","relatedCode":"AM-001"}' "$review_code" "$now")
review_created=$(curl -fsS -X POST "http://127.0.0.1:${BACKEND_PORT}/api/reviews" -H "Authorization: Bearer $token" -H 'Content-Type: application/json' -d "$review_payload")
review_id=$(printf '%s' "$review_created" | jq -er '.data.id')
review_version=$(printf '%s' "$review_created" | jq -er '.data.version')
direct_sign=$(printf '{"status":"signed","expectedVersion":%s,"reason":"must not bypass peer review"}' "$review_version")
direct_status=$(curl -sS -o /dev/null -w '%{http_code}' -X POST "http://127.0.0.1:${BACKEND_PORT}/api/reviews/$review_id/transition" -H "Authorization: Bearer $token" -H 'Content-Type: application/json' -d "$direct_sign")
[ "$direct_status" = "422" ]
peer_request=$(printf '{"status":"peer_review","expectedVersion":%s,"reason":"submit for independent review"}' "$review_version")
peer_result=$(curl -fsS -X POST "http://127.0.0.1:${BACKEND_PORT}/api/reviews/$review_id/transition" -H "Authorization: Bearer $token" -H 'Content-Type: application/json' -d "$peer_request")
peer_version=$(printf '%s' "$peer_result" | jq -er '.data.version')
sign_request=$(printf '{"status":"signed","expectedVersion":%s,"reason":"independent result approval"}' "$peer_version")
self_sign_status=$(curl -sS -o /dev/null -w '%{http_code}' -X POST "http://127.0.0.1:${BACKEND_PORT}/api/reviews/$review_id/transition" -H "Authorization: Bearer $token" -H 'Content-Type: application/json' -d "$sign_request")
[ "$self_sign_status" = "422" ]
operator_sign_status=$(curl -sS -o /dev/null -w '%{http_code}' -X POST "http://127.0.0.1:${BACKEND_PORT}/api/reviews/$review_id/transition" -H "Authorization: Bearer $operator_token" -H 'Content-Type: application/json' -d "$sign_request")
[ "$operator_sign_status" = "403" ]
curl -fsS -X POST "http://127.0.0.1:${BACKEND_PORT}/api/reviews/$review_id/transition" -H "Authorization: Bearer $reviewer_token" -H 'Content-Type: application/json' -d "$sign_request" | jq -e '.data.status == "signed" and .data.reviewRequestedBy == "admin" and .data.peerReviewedBy == "reviewer" and .data.signedBy == "reviewer"' >/dev/null
curl -fsS "http://127.0.0.1:${BACKEND_PORT}/api/audits?page=1&pageSize=100" -H "Authorization: Bearer $token" | jq -e '.meta.total >= 2' >/dev/null
curl -fsS "http://127.0.0.1:${BACKEND_PORT}/api/audit-summary?windowHours=24" -H "Authorization: Bearer $token" | jq -e '.data.total >= 2 and .data.transitions >= 1' >/dev/null
docker compose ps
[ "${KEEP_RUNNING:-0}" = "1" ] && echo "KEEP_RUNNING=1: containers left running for browser validation"
