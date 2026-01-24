package auth

import (
	"context"
	"fmt"
	"time"

	identra_v1 "github.com/poly-workshop/identra/gen/go/identra/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// IdentraClient wraps the gRPC client for Identra service
type IdentraClient struct {
	client identra_v1.IdentraServiceClient
	conn   *grpc.ClientConn
}

// NewIdentraClient creates a new Identra gRPC client
func NewIdentraClient(endpoint string) (*IdentraClient, error) {
	// TODO: Add support for TLS credentials in production
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Identra: %w", err)
	}

	return &IdentraClient{
		client: identra_v1.NewIdentraServiceClient(conn),
		conn:   conn,
	}, nil
}

// GetJWKS fetches the JSON Web Key Set from Identra
func (c *IdentraClient) GetJWKS(ctx context.Context) (*identra_v1.GetJWKSResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := c.client.GetJWKS(ctx, &identra_v1.GetJWKSRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get JWKS: %w", err)
	}

	return resp, nil
}

// Close closes the gRPC connection
func (c *IdentraClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
