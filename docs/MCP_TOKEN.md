# MCP Token Authentication

This document describes the MCP (Model Context Protocol) Token authentication feature in slips-core.

## Overview

MCP Tokens provide an alternative authentication method to JWT tokens. They are UUID-based tokens that can be created by authenticated users to enable programmatic access to the Task and Tag services.

## Key Features

- **UUID-based tokens**: Each MCP token is a unique UUID that provides secure, non-expiring (or time-limited) access
- **User-scoped**: Tokens are owned by users and can only access that user's resources
- **Named tokens**: Each token has a human-readable name for easy identification
- **Revocable**: Tokens can be revoked at any time without affecting the user's JWT authentication
- **Optional expiration**: Tokens can be created with an expiration time or without (never expires)
- **Usage tracking**: The system tracks when each token was last used

## Architecture

### Database Schema

The `mcp_tokens` table stores all MCP tokens:

```sql
CREATE TABLE mcp_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token UUID UNIQUE NOT NULL,                    -- The actual token value
    user_id VARCHAR(255) NOT NULL,                 -- Owner of the token
    name VARCHAR(255) NOT NULL,                     -- Human-readable name
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,                          -- Optional expiration
    last_used_at TIMESTAMP,                        -- Last usage timestamp
    is_active BOOLEAN NOT NULL DEFAULT TRUE        -- Active/revoked status
);
```

### Components

1. **Domain Layer** (`internal/mcptoken/domain/`)
   - `MCPToken`: Entity representing an MCP token
   - `Repository`: Interface for token persistence operations

2. **Application Layer** (`internal/mcptoken/application/`)
   - `Service`: Business logic for token management
   - Methods: CreateToken, GetToken, ListTokens, RevokeToken, DeleteToken, ValidateToken

3. **Infrastructure Layer** (`internal/mcptoken/infra/`)
   - `postgres/`: PostgreSQL repository implementation
   - `grpc/`: gRPC server implementation

4. **Auth Package** (`pkg/auth/`)
   - `mcptoken.go`: MCP token validation and extraction logic
   - `interceptor.go`: Updated to support both JWT and MCP token authentication

## Usage

### Creating an MCP Token

Users must first authenticate with a JWT token, then can create MCP tokens:

```bash
# Authenticate with JWT
grpcurl -H "Authorization: Bearer <jwt-token>" \
  -d '{
    "name": "My API Token",
    "expires_at": "2026-12-31T23:59:59Z"
  }' \
  localhost:9090 mcptoken.v1.MCPTokenService/CreateMCPToken
```

Response:

```json
{
  "token": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "token": "98765432-e89b-12d3-a456-426614174000",
    "name": "My API Token",
    "created_at": "2026-01-24T10:00:00Z",
    "expires_at": "2026-12-31T23:59:59Z",
    "is_active": true
  }
}
```

**Important**: Save the `token` UUID value immediately - it cannot be retrieved again later.

### Using an MCP Token

Use the token in the Authorization header with the `MCP-Token` scheme:

```bash
# Access Task service with MCP token
grpcurl -H "Authorization: MCP-Token 98765432-e89b-12d3-a456-426614174000" \
  -d '{"title": "My Task", "notes": "Task notes"}' \
  localhost:9090 task.v1.TaskService/CreateTask
```

The same token works for all services:

```bash
# Access Tag service with MCP token
grpcurl -H "Authorization: MCP-Token 98765432-e89b-12d3-a456-426614174000" \
  -d '{"name": "important"}' \
  localhost:9090 tag.v1.TagService/CreateTag
```

### Listing Tokens

```bash
grpcurl -H "Authorization: Bearer <jwt-token>" \
  localhost:9090 mcptoken.v1.MCPTokenService/ListMCPTokens
```

### Revoking a Token

```bash
grpcurl -H "Authorization: Bearer <jwt-token>" \
  -d '{"id": "123e4567-e89b-12d3-a456-426614174000"}' \
  localhost:9090 mcptoken.v1.MCPTokenService/RevokeMCPToken
```

### Deleting a Token

```bash
grpcurl -H "Authorization: Bearer <jwt-token>" \
  -d '{"id": "123e4567-e89b-12d3-a456-426614174000"}' \
  localhost:9090 mcptoken.v1.MCPTokenService/DeleteMCPToken
```

## Authentication Flow

### JWT Authentication Flow

1. Client sends request with `Authorization: Bearer <jwt-token>`
2. Auth interceptor extracts JWT token
3. JWT validator validates token with JWKS
4. User ID is extracted from JWT claims
5. User ID is added to request context
6. Request proceeds to service handler

### MCP Token Authentication Flow

1. Client sends request with `Authorization: MCP-Token <uuid>`
2. Auth interceptor detects MCP-Token scheme
3. MCP token validator retrieves token from database
4. Token is validated (active and not expired)
5. User ID associated with token is extracted
6. Last used timestamp is updated asynchronously
7. User ID is added to request context
8. Request proceeds to service handler

## Security Considerations

1. **Token Storage**: MCP tokens should be stored securely by clients (e.g., environment variables, secret management systems)
2. **Token Rotation**: Users should periodically rotate tokens and revoke unused ones
3. **Expiration**: Consider setting expiration times for tokens used in automated systems
4. **Naming**: Use descriptive names to identify where tokens are used
5. **Monitoring**: The `last_used_at` field helps identify stale or unused tokens

## Migration

To enable MCP token support in an existing deployment:

1. Run the migration:

   ```bash
   make migrate-up
   ```

2. Restart the service to load the new code

3. Users can immediately start creating MCP tokens through the API

## API Reference

### MCPTokenService

#### CreateMCPToken

Creates a new MCP token for the authenticated user.

**Request:**

```protobuf
message CreateMCPTokenRequest {
  string name = 1;                                    // Required
  google.protobuf.Timestamp expires_at = 2;          // Optional
}
```

**Response:**

```protobuf
message CreateMCPTokenResponse {
  MCPToken token = 1;
}
```

#### GetMCPToken

Retrieves an MCP token by ID (user must own the token).

**Request:**

```protobuf
message GetMCPTokenRequest {
  string id = 1;
}
```

**Response:**

```protobuf
message GetMCPTokenResponse {
  MCPToken token = 1;
}
```

#### ListMCPTokens

Lists all MCP tokens for the authenticated user.

**Request:**

```protobuf
message ListMCPTokensRequest {}
```

**Response:**

```protobuf
message ListMCPTokensResponse {
  repeated MCPToken tokens = 1;
}
```

#### RevokeMCPToken

Revokes (deactivates) an MCP token.

**Request:**

```protobuf
message RevokeMCPTokenRequest {
  string id = 1;
}
```

**Response:**

```protobuf
message RevokeMCPTokenResponse {}
```

#### DeleteMCPToken

Permanently deletes an MCP token.

**Request:**

```protobuf
message DeleteMCPTokenRequest {
  string id = 1;
}
```

**Response:**

```protobuf
message DeleteMCPTokenResponse {}
```

## Implementation Details

### Token Generation

- Tokens are generated using `uuid.New()` which creates a random UUID v4
- Each token is unique and stored in the database before being returned

### Token Validation

- Validation checks if the token exists, is active, and not expired
- Expired or inactive tokens are rejected immediately
- Last used timestamp is updated asynchronously to avoid blocking the request

### Dual Authentication Support

- The auth interceptor (`UnaryServerInterceptorWithMCP`) supports both JWT and MCP tokens
- Token type is determined by the Authorization header prefix:
  - `Bearer <token>` → JWT authentication
  - `MCP-Token <uuid>` → MCP token authentication
- Both authentication methods result in the same user context being created

## Troubleshooting

### "invalid MCP token format" Error

- Ensure the Authorization header uses the correct format: `MCP-Token <uuid>`
- Verify the UUID is valid (36 characters with hyphens)

### "invalid MCP token" Error

- Token may be revoked (check `is_active` status)
- Token may be expired (check `expires_at`)
- Token may not exist in the database

### "unauthorized: user mismatch" Error

- You're trying to access a token that belongs to a different user
- Each user can only manage their own tokens

### Token Not Working After Creation

- Ensure you're using the `token` field from the response, not the `id` field
- The `id` is for management operations, the `token` is for authentication
