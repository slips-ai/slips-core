package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	mcptokenv1 "github.com/slips-ai/slips-core/gen/api/proto/mcptoken/v1"
	"github.com/slips-ai/slips-core/internal/mcptoken/application"
	"github.com/slips-ai/slips-core/internal/mcptoken/domain"
	"github.com/slips-ai/slips-core/pkg/grpcerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MCPTokenServer implements the MCPTokenService gRPC server
type MCPTokenServer struct {
	mcptokenv1.UnimplementedMCPTokenServiceServer
	service *application.Service
}

// NewMCPTokenServer creates a new MCP token gRPC server
func NewMCPTokenServer(service *application.Service) *MCPTokenServer {
	return &MCPTokenServer{
		service: service,
	}
}

// CreateMCPToken creates a new MCP token
func (s *MCPTokenServer) CreateMCPToken(ctx context.Context, req *mcptokenv1.CreateMCPTokenRequest) (*mcptokenv1.CreateMCPTokenResponse, error) {
	// Validate input
	if err := grpcerrors.ValidateNotEmpty(req.Name, "name"); err != nil {
		return nil, err
	}
	if err := grpcerrors.ValidateLength(req.Name, "name", 255); err != nil {
		return nil, err
	}

	// Convert protobuf timestamp to *time.Time
	var expiresAt *time.Time
	if req.ExpiresAt != nil && req.ExpiresAt.IsValid() {
		t := req.ExpiresAt.AsTime()
		expiresAt = &t
	}

	token, err := s.service.CreateToken(ctx, req.Name, expiresAt)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to create MCP token")
	}

	return &mcptokenv1.CreateMCPTokenResponse{
		Token: s.toProto(token),
	}, nil
}

// GetMCPToken retrieves an MCP token by ID
func (s *MCPTokenServer) GetMCPToken(ctx context.Context, req *mcptokenv1.GetMCPTokenRequest) (*mcptokenv1.GetMCPTokenResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid token ID format")
	}

	token, err := s.service.GetToken(ctx, id)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to get MCP token")
	}

	return &mcptokenv1.GetMCPTokenResponse{
		Token: s.toProto(token),
	}, nil
}

// ListMCPTokens retrieves all MCP tokens for the authenticated user
func (s *MCPTokenServer) ListMCPTokens(ctx context.Context, req *mcptokenv1.ListMCPTokensRequest) (*mcptokenv1.ListMCPTokensResponse, error) {
	tokens, err := s.service.ListTokens(ctx)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to list MCP tokens")
	}

	protoTokens := make([]*mcptokenv1.MCPToken, len(tokens))
	for i, token := range tokens {
		protoTokens[i] = s.toProto(token)
	}

	return &mcptokenv1.ListMCPTokensResponse{
		Tokens: protoTokens,
	}, nil
}

// RevokeMCPToken revokes an MCP token
func (s *MCPTokenServer) RevokeMCPToken(ctx context.Context, req *mcptokenv1.RevokeMCPTokenRequest) (*mcptokenv1.RevokeMCPTokenResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid token ID format")
	}

	if err := s.service.RevokeToken(ctx, id); err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to revoke MCP token")
	}

	return &mcptokenv1.RevokeMCPTokenResponse{}, nil
}

// DeleteMCPToken deletes an MCP token
func (s *MCPTokenServer) DeleteMCPToken(ctx context.Context, req *mcptokenv1.DeleteMCPTokenRequest) (*mcptokenv1.DeleteMCPTokenResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid token ID format")
	}

	if err := s.service.DeleteToken(ctx, id); err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to delete MCP token")
	}

	return &mcptokenv1.DeleteMCPTokenResponse{}, nil
}

// Helper function to convert domain model to proto
func (s *MCPTokenServer) toProto(token *domain.MCPToken) *mcptokenv1.MCPToken {
	protoToken := &mcptokenv1.MCPToken{
		Id:        token.ID.String(),
		Token:     token.Token.String(),
		Name:      token.Name,
		CreatedAt: timestamppb.New(token.CreatedAt),
		IsActive:  token.IsActive,
	}

	if token.ExpiresAt != nil {
		protoToken.ExpiresAt = timestamppb.New(*token.ExpiresAt)
	}

	if token.LastUsedAt != nil {
		protoToken.LastUsedAt = timestamppb.New(*token.LastUsedAt)
	}

	return protoToken
}
