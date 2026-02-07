package grpc

import (
	"context"

	"github.com/google/uuid"
	tagv1 "github.com/slips-ai/slips-core/gen/go/tag/v1"
	"github.com/slips-ai/slips-core/internal/tag/application"
	"github.com/slips-ai/slips-core/pkg/grpcerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TagServer implements the TagService gRPC server
type TagServer struct {
	tagv1.UnimplementedTagServiceServer
	service *application.Service
}

// NewTagServer creates a new tag gRPC server
func NewTagServer(service *application.Service) *TagServer {
	return &TagServer{
		service: service,
	}
}

// CreateTag creates a new tag
func (s *TagServer) CreateTag(ctx context.Context, req *tagv1.CreateTagRequest) (*tagv1.CreateTagResponse, error) {
	// Validate input
	if err := grpcerrors.ValidateTagName(req.Name); err != nil {
		return nil, err
	}

	tag, err := s.service.CreateTag(ctx, req.Name)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to create tag")
	}

	return &tagv1.CreateTagResponse{
		Tag: &tagv1.Tag{
			Id:        tag.ID.String(),
			Name:      tag.Name,
			CreatedAt: timestamppb.New(tag.CreatedAt),
			UpdatedAt: timestamppb.New(tag.UpdatedAt),
		},
	}, nil
}

// GetTag retrieves a tag by ID
func (s *TagServer) GetTag(ctx context.Context, req *tagv1.GetTagRequest) (*tagv1.GetTagResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tag ID format")
	}

	tag, err := s.service.GetTag(ctx, id)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to get tag")
	}

	return &tagv1.GetTagResponse{
		Tag: &tagv1.Tag{
			Id:        tag.ID.String(),
			Name:      tag.Name,
			CreatedAt: timestamppb.New(tag.CreatedAt),
			UpdatedAt: timestamppb.New(tag.UpdatedAt),
		},
	}, nil
}

// UpdateTag updates a tag
func (s *TagServer) UpdateTag(ctx context.Context, req *tagv1.UpdateTagRequest) (*tagv1.UpdateTagResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tag ID format")
	}

	// Validate input
	if err := grpcerrors.ValidateTagName(req.Name); err != nil {
		return nil, err
	}

	tag, err := s.service.UpdateTag(ctx, id, req.Name)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to update tag")
	}

	return &tagv1.UpdateTagResponse{
		Tag: &tagv1.Tag{
			Id:        tag.ID.String(),
			Name:      tag.Name,
			CreatedAt: timestamppb.New(tag.CreatedAt),
			UpdatedAt: timestamppb.New(tag.UpdatedAt),
		},
	}, nil
}

// DeleteTag deletes a tag
func (s *TagServer) DeleteTag(ctx context.Context, req *tagv1.DeleteTagRequest) (*tagv1.DeleteTagResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tag ID format")
	}

	if err := s.service.DeleteTag(ctx, id); err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to delete tag")
	}

	return &tagv1.DeleteTagResponse{}, nil
}

// ListTags lists tags with pagination
func (s *TagServer) ListTags(ctx context.Context, req *tagv1.ListTagsRequest) (*tagv1.ListTagsResponse, error) {
	// Reject page_token if provided (not yet implemented)
	if req.PageToken != "" {
		return nil, status.Errorf(codes.Unimplemented, "page_token is not supported yet")
	}

	pageSize := int(req.PageSize)
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 30
	}

	// Always return the first page (offset 0) until pagination tokens are implemented
	offset := 0

	// Validate int32 bounds at gRPC layer before calling repository
	if err := grpcerrors.ValidateInt32Range(pageSize, "limit"); err != nil {
		return nil, err
	}
	if err := grpcerrors.ValidateInt32Range(offset, "offset"); err != nil {
		return nil, err
	}

	tags, err := s.service.ListTags(ctx, pageSize, offset)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to list tags")
	}

	protoTags := make([]*tagv1.Tag, len(tags))
	for i, tag := range tags {
		protoTags[i] = &tagv1.Tag{
			Id:        tag.ID.String(),
			Name:      tag.Name,
			CreatedAt: timestamppb.New(tag.CreatedAt),
			UpdatedAt: timestamppb.New(tag.UpdatedAt),
		}
	}

	// Note: next_page_token is not implemented yet
	// Future implementation would return a token when len(tags) == pageSize
	return &tagv1.ListTagsResponse{
		Tags: protoTags,
	}, nil
}
