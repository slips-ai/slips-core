# Identra JWKS gRPC Integration

This document describes the changes made to integrate Identra's JWKS endpoint via gRPC instead of HTTP.

## Summary

The slips-core application now accesses Identra's JWKS (JSON Web Key Set) endpoint through gRPC instead of HTTP. This provides better performance, type safety, and consistency with the rest of the service architecture.

## Changes Made

### 1. Dependencies

- Added `github.com/poly-workshop/identra v0.1.2` to provide the generated gRPC client code

### 2. Configuration Updates

**File: `pkg/config/config.go`**

- Changed `JWKSEndpoint` (HTTP URL) to `IdentraGRPCEndpoint` (gRPC host:port)
- Updated default from `http://localhost:8080/.well-known/jwks.json` to `localhost:8080`

**File: `config.yaml`**

- Updated `auth.jwks_endpoint` to `auth.identra_grpc_endpoint`
- Changed value from HTTP URL to gRPC endpoint `localhost:8080`

### 3. New Files

**File: `pkg/auth/identra_client.go`**

- Created `IdentraClient` struct to wrap the gRPC client
- Implements `NewIdentraClient(endpoint string)` to create a new client
- Implements `GetJWKS(ctx)` to fetch JWKS via gRPC
- Implements `Close()` for cleanup
- Currently uses insecure credentials (TODO: add TLS support for production)

### 4. Updated Files

**File: `pkg/auth/jwt.go`**

- Removed HTTP-specific imports (`encoding/json`, `io`, `net/http`)
- Removed `JWKS` and `JWKSKey` structs (now use Identra proto types)
- Updated `JWTValidator` struct to use `*IdentraClient` instead of HTTP client
- Updated `NewJWTValidator()` to accept `*IdentraClient` instead of URL
- Updated `FetchJWKS()` to call gRPC endpoint instead of HTTP
- Parses keys from `identra_v1.GetJWKSResponse.Keys` (proto generated types)

**File: `cmd/server/main.go`**

- Added Identra client initialization before JWT validator
- Updated JWT validator creation to use gRPC client
- Added proper cleanup with `defer identraClient.Close()`

**File: `pkg/auth/jwt_test.go`**

- Created new simplified test file (old tests backed up as `jwt_test.go.old`)
- Tests now focus on helper functions that don't require HTTP mocking
- Integration tests for JWKS fetching should be added later with gRPC mocking

## Usage

The server now connects to Identra via gRPC on startup:

```go
// Initialize Identra gRPC client
identraClient, err := auth.NewIdentraClient(cfg.Auth.IdentraGRPCEndpoint)
if err != nil {
    log.Fatal(err)
}
defer identraClient.Close()

// Initialize JWT validator with gRPC client
jwtValidator := auth.NewJWTValidator(identraClient, cfg.Auth.ExpectedIssuer)

// Fetch JWKS keys
if err := jwtValidator.FetchJWKS(ctx); err != nil {
    log.Fatal(err)
}
```

## Configuration Example

```yaml
auth:
  identra_grpc_endpoint: localhost:8080  # Identra gRPC server
  expected_issuer: identra
```

## Protocol Buffer Types

The integration uses these types from `github.com/poly-workshop/identra/gen/go/identra/v1`:

- `IdentraServiceClient` - gRPC client interface
- `GetJWKSRequest` - Request message (empty)
- `GetJWKSResponse` - Response containing `repeated JSONWebKey keys`
- `JSONWebKey` - Individual key with fields: `kty`, `alg`, `use`, `kid`, `n`, `e`

## TODO

1. Add TLS support for production deployments (currently uses insecure credentials)
2. Implement gRPC-based integration tests
3. Add periodic JWKS refresh mechanism for key rotation
4. Consider caching strategy for JWKS responses

## Testing

Run tests with:

```bash
go test ./pkg/auth/...
```

Build the server:

```bash
go build -o bin/slips-core cmd/server/main.go
```

## Benefits

1. **Type Safety**: Using protobuf-generated types instead of JSON parsing
2. **Performance**: gRPC binary protocol is more efficient than HTTP/JSON
3. **Consistency**: All service communication now uses gRPC
4. **Maintainability**: Shared proto definitions ensure API compatibility
