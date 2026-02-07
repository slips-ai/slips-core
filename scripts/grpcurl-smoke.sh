#!/usr/bin/env bash

set -euo pipefail

CORE_ADDR="${CORE_ADDR:-localhost:9090}"
POSTGRES_CONTAINER="${POSTGRES_CONTAINER:-postgres}"
POSTGRES_DB="${POSTGRES_DB:-slips}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"

have_cmd() {
  command -v "$1" >/dev/null 2>&1
}

require_cmd() {
  if ! have_cmd "$1"; then
    echo "error: missing required command: $1" >&2
    exit 1
  fi
}

gen_uuid() {
  if [[ -r /proc/sys/kernel/random/uuid ]]; then
    cat /proc/sys/kernel/random/uuid
    return
  fi

  if have_cmd uuidgen; then
    uuidgen
    return
  fi

  # Fallback: Python (common in dev envs)
  python3 - <<'PY'
import uuid
print(uuid.uuid4())
PY
}

AUTH_HEADER="${SLIPS_AUTH_HEADER:-}"

if [[ -z "$AUTH_HEADER" ]]; then
  if [[ -n "${SLIPS_JWT:-}" ]]; then
    AUTH_HEADER="Bearer ${SLIPS_JWT}"
  elif [[ -n "${SLIPS_MCP_TOKEN:-}" ]]; then
    AUTH_HEADER="MCP-Token ${SLIPS_MCP_TOKEN}"
  fi
fi

seed_mcp_token() {
  local user_id token id

  require_cmd docker

  user_id="${SLIPS_SEED_USER_ID:-grpcurl-dev-user}"
  token="$(gen_uuid)"
  id="$(gen_uuid)"

  echo "Seeding MCP token into Postgres (user_id=${user_id})..." >&2
  docker exec -i "${POSTGRES_CONTAINER}" psql \
    -U "${POSTGRES_USER}" \
    -d "${POSTGRES_DB}" \
    -v ON_ERROR_STOP=1 \
    -c "INSERT INTO mcp_tokens (id, token, user_id, name, is_active) VALUES ('${id}', '${token}', '${user_id}', 'grpcurl smoke', TRUE);" \
    >/dev/null

  AUTH_HEADER="MCP-Token ${token}"
  export SLIPS_MCP_TOKEN="${token}"
}

main() {
  require_cmd grpcurl

  echo "== grpcurl: listing services on ${CORE_ADDR} ==" >&2
  grpcurl -plaintext "${CORE_ADDR}" list

  if [[ -z "$AUTH_HEADER" ]]; then
    if [[ "${SEED_MCP_TOKEN:-}" == "1" ]]; then
      seed_mcp_token
    else
      cat >&2 <<EOF

No auth provided.

To run authenticated calls, set one of:
  - SLIPS_JWT=<jwt>         (uses Authorization: Bearer ...)
  - SLIPS_MCP_TOKEN=<uuid>  (uses Authorization: MCP-Token ...)
  - SLIPS_AUTH_HEADER='Bearer ...' or 'MCP-Token ...'

Or to auto-seed an MCP token into the local docker Postgres (dev only):
  SEED_MCP_TOKEN=1 make grpcurl-smoke
EOF
      exit 0
    fi
  fi

  echo "== grpcurl: TaskService smoke (create + list) ==" >&2
  grpcurl -plaintext \
    -H "Authorization: ${AUTH_HEADER}" \
    -d '{"title":"grpcurl smoke task","notes":"created by scripts/grpcurl-smoke.sh"}' \
    "${CORE_ADDR}" task.v1.TaskService/CreateTask

  grpcurl -plaintext \
    -H "Authorization: ${AUTH_HEADER}" \
    -d '{"page_size":10,"page_token":""}' \
    "${CORE_ADDR}" task.v1.TaskService/ListTasks

  echo "== grpcurl: TagService smoke (create + list) ==" >&2
  grpcurl -plaintext \
    -H "Authorization: ${AUTH_HEADER}" \
    -d '{"name":"grpcurl-smoke"}' \
    "${CORE_ADDR}" tag.v1.TagService/CreateTag || true

  grpcurl -plaintext \
    -H "Authorization: ${AUTH_HEADER}" \
    -d '{"page_size":10,"page_token":""}' \
    "${CORE_ADDR}" tag.v1.TagService/ListTags
}

main "$@"
