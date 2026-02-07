# grpcurl API Testing (slips-core)

This document shows how to test slips-core gRPC APIs using `grpcurl`.

## Prerequisites

- `slips-core` running locally (default gRPC address: `localhost:9090`)
- `grpcurl` installed

Server reflection is enabled in slips-core, so `grpcurl list/describe` works without `.proto` files.

## Start dependencies + run slips-core

From the slips-core folder:

```bash
make docker-up
make db-create
make migrate-up
make run
```

## Discover services and methods

```bash
grpcurl -plaintext localhost:9090 list
grpcurl -plaintext localhost:9090 describe
grpcurl -plaintext localhost:9090 describe task.v1.TaskService
```

## Authentication header

Most methods require an `Authorization` metadata header:

- JWT: `Authorization: Bearer <jwt>`
- MCP token: `Authorization: MCP-Token <uuid>`

Public (no-auth) endpoints:

- `auth.v1.AuthService/GetAuthorizationURL`
- `auth.v1.AuthService/HandleCallback`
- `auth.v1.AuthService/RefreshToken`

## Option A: Use an existing JWT

If you already have a valid Identra JWT access token:

```bash
export SLIPS_JWT='<your-access-jwt>'

grpcurl -plaintext \
  -H "Authorization: Bearer ${SLIPS_JWT}" \
  -d '{"title":"hello","notes":"from grpcurl"}' \
  localhost:9090 task.v1.TaskService/CreateTask
```

## Option B (dev-only): Seed an MCP token in Postgres (no OAuth needed)

This is the quickest way to run authenticated API calls locally.

1) Generate UUIDs:

```bash
TOKEN_ID=$(cat /proc/sys/kernel/random/uuid)
TOKEN=$(cat /proc/sys/kernel/random/uuid)
```

1) Insert a token row into the local docker Postgres:

```bash
docker exec -i postgres psql -U postgres -d slips -v ON_ERROR_STOP=1 \
  -c "INSERT INTO mcp_tokens (id, token, user_id, name, is_active) VALUES ('${TOKEN_ID}', '${TOKEN}', 'grpcurl-dev-user', 'grpcurl seed', TRUE);"
```

1) Use it with grpcurl:

```bash
export SLIPS_MCP_TOKEN="${TOKEN}"

grpcurl -plaintext \
  -H "Authorization: MCP-Token ${SLIPS_MCP_TOKEN}" \
  -d '{"title":"mcp authed","notes":"ok"}' \
  localhost:9090 task.v1.TaskService/CreateTask
```

## Common calls

### Task

```bash
grpcurl -plaintext \
  -H "Authorization: MCP-Token ${SLIPS_MCP_TOKEN}" \
  -d '{"page_size":10,"page_token":""}' \
  localhost:9090 task.v1.TaskService/ListTasks
```

### Tag

```bash
grpcurl -plaintext \
  -H "Authorization: MCP-Token ${SLIPS_MCP_TOKEN}" \
  -d '{"name":"important"}' \
  localhost:9090 tag.v1.TagService/CreateTag

grpcurl -plaintext \
  -H "Authorization: MCP-Token ${SLIPS_MCP_TOKEN}" \
  -d '{"page_size":10,"page_token":""}' \
  localhost:9090 tag.v1.TagService/ListTags
```

### MCP Token service (requires auth)

```bash
grpcurl -plaintext \
  -H "Authorization: Bearer ${SLIPS_JWT}" \
  -d '{"name":"My API Token"}' \
  localhost:9090 mcptoken.v1.MCPTokenService/CreateMCPToken
```

## One-command smoke test

There is a helper script that exercises `CreateTask/ListTasks` and `CreateTag/ListTags`.

With an existing JWT:

```bash
SLIPS_JWT='<jwt>' make grpcurl-smoke
```

Or auto-seed an MCP token into the local docker Postgres (dev-only):

```bash
SEED_MCP_TOKEN=1 make grpcurl-smoke
```
