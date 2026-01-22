# JWT Authentication and User-Scoped Resources

This document describes the JWT authentication and user-scoped resource implementation in slips-core.

## Overview

Tasks and Tags in slips-core are now user-scoped, meaning each user can only access their own data. Authentication is performed using JWT tokens from Identra, an identity provider.

## Key Components

### 1. Database Schema

Both `tasks` and `tags` tables include an `owner_id` column:
- Type: VARCHAR(255)
- NOT NULL
- Indexed for performance

Migration: `migrations/002_add_owner_id.up.sql`

### 2. JWT Validator (`pkg/auth/jwt.go`)

The JWT validator:
- Fetches JWKS (JSON Web Key Set) from Identra
- Validates JWT tokens using RSA public keys
- Verifies token type is "access" (rejects "refresh" tokens)
- Verifies issuer matches expected Identra instance
- Extracts user ID from `sub` claim (or `uid` for compatibility)

### 3. gRPC Interceptor (`pkg/auth/interceptor.go`)

The auth interceptor:
- Runs on every gRPC request
- Extracts JWT from `Authorization: Bearer <token>` header
- Validates the token
- Injects user ID into request context

### 4. Context Helpers (`pkg/auth/context.go`)

Helper functions for managing user ID in context:
- `WithUserID(ctx, userID)` - Add user ID to context
- `GetUserID(ctx)` - Extract user ID from context

## Configuration

Add to `config.yaml` or set via environment variables:

```yaml
auth:
  jwks_endpoint: http://localhost:8080/.well-known/jwks.json
  expected_issuer: http://localhost:8080
```

Environment variables (with `SLIPS_` prefix):
- `SLIPS_AUTH_JWKS_ENDPOINT`
- `SLIPS_AUTH_EXPECTED_ISSUER`

## Token Requirements

Valid tokens must:
1. Be signed with RSA using a key from the JWKS
2. Have `typ` claim set to "access"
3. Have `iss` claim matching `expected_issuer`
4. Not be expired
5. Contain `sub` or `uid` claim with user ID

## User Isolation

All operations are scoped to the authenticated user:

### Create Operations
- Extract user ID from context
- Set `owner_id` on new resource

### Read Operations (Get/List)
- Filter by `owner_id = authenticated_user_id`
- Returns 404/empty if resource doesn't exist or belongs to another user

### Update Operations
- Fetch resource with `owner_id` check
- Returns error if resource doesn't exist or belongs to another user

### Delete Operations
- Delete only if `owner_id` matches authenticated user
- Silent failure if resource doesn't exist or belongs to another user

## Error Handling

### Authentication Errors (codes.Unauthenticated)
- Missing authorization header
- Invalid token format
- Expired token
- Invalid signature
- Wrong token type (refresh instead of access)
- Invalid issuer

### Authorization Errors (codes.NotFound/PermissionDenied)
- Resource not found (may belong to another user)
- Implicit: operations fail silently for other users' resources

## Security Considerations

1. **JWKS Refresh**: Currently JWKS is fetched once at startup. In production, implement periodic refresh.
2. **Token Expiration**: Tokens are validated for expiration. Clients must refresh tokens.
3. **HTTPS Required**: In production, use HTTPS for JWKS endpoint and gRPC.
4. **Owner ID Immutability**: Owner ID cannot be changed after resource creation.

## Testing

### Unit Tests
- `pkg/auth/context_test.go` - Context helpers
- `pkg/auth/jwt_test.go` - Token extraction and user ID extraction

### Integration Testing
Use a valid JWT token from Identra:

```bash
# Example gRPC call with authentication
grpcurl -H "Authorization: Bearer <your-jwt-token>" \
  -d '{"title": "My Task", "notes": "Test"}' \
  localhost:9090 task.v1.TaskService/CreateTask
```

## Migration Notes

### Applying Migrations

```bash
# Set database password
export DB_PASSWORD=your_password

# Run migrations
make migrate-up
```

### Existing Data

The migration sets `owner_id` to empty string by default. You may want to:
1. Delete all existing data before migration
2. Or manually assign ownership after migration

## Future Enhancements

1. **JWKS Caching & Refresh**: Implement periodic JWKS refresh
2. **Token Revocation**: Check token revocation list
3. **Multi-tenancy**: Add organization/tenant scoping
4. **Admin Roles**: Allow admins to access all resources
5. **Shared Resources**: Allow resource sharing between users
