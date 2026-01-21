package grpc

import (
	"context"

	"github.com/google/uuid"
	tagv1 "github.com/slips-ai/slips-core/gen/proto/tag/v1"
	"github.com/slips-ai/slips-core/features/tag/application"
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
	tag, err := s.service.CreateTag(ctx, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create tag: %v", err)
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid tag ID: %v", err)
	}

	tag, err := s.service.GetTag(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "tag not found: %v", err)
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid tag ID: %v", err)
	}

	tag, err := s.service.UpdateTag(ctx, id, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update tag: %v", err)
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid tag ID: %v", err)
	}

	if err := s.service.DeleteTag(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete tag: %v", err)
	}

	return &tagv1.DeleteTagResponse{}, nil
}

// ListTags lists tags with pagination
func (s *TagServer) ListTags(ctx context.Context, req *tagv1.ListTagsRequest) (*tagv1.ListTagsResponse, error) {
	pageSize := int(req.PageSize)
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 30
	}

	offset := 0
	if req.PageToken != "" {
		// In a real implementation, decode the page token
		// For simplicity, we'll skip this for now
	}

	tags, err := s.service.ListTags(ctx, pageSize, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list tags: %v", err)
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

	return &tagv1.ListTagsResponse{
		Tags: protoTags,
	}, nil
}
