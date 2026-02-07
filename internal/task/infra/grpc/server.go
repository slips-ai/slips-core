package grpc

import (
	"context"

	"github.com/google/uuid"
	taskv1 "github.com/slips-ai/slips-core/gen/go/task/v1"
	"github.com/slips-ai/slips-core/internal/task/application"
	"github.com/slips-ai/slips-core/internal/task/domain"
	"github.com/slips-ai/slips-core/pkg/grpcerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TaskServer implements the TaskService gRPC server
type TaskServer struct {
	taskv1.UnimplementedTaskServiceServer
	service *application.Service
}

// NewTaskServer creates a new task gRPC server
func NewTaskServer(service *application.Service) *TaskServer {
	return &TaskServer{
		service: service,
	}
}

// CreateTask creates a new task
func (s *TaskServer) CreateTask(ctx context.Context, req *taskv1.CreateTaskRequest) (*taskv1.CreateTaskResponse, error) {
	// Validate input
	if err := grpcerrors.ValidateNotEmpty(req.Title, "title"); err != nil {
		return nil, err
	}
	if err := grpcerrors.ValidateLength(req.Title, "title", grpcerrors.MaxTitleLength); err != nil {
		return nil, err
	}
	if err := grpcerrors.ValidateLength(req.Notes, "notes", grpcerrors.MaxNotesLength); err != nil {
		return nil, err
	}

	task, err := s.service.CreateTask(ctx, req.Title, req.Notes, req.TagNames)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to create task")
	}

	return &taskv1.CreateTaskResponse{
		Task: taskToProto(task),
	}, nil
}

// GetTask retrieves a task by ID
func (s *TaskServer) GetTask(ctx context.Context, req *taskv1.GetTaskRequest) (*taskv1.GetTaskResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid task ID format")
	}

	task, err := s.service.GetTask(ctx, id)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to get task")
	}

	return &taskv1.GetTaskResponse{
		Task: taskToProto(task),
	}, nil
}

// UpdateTask updates a task
func (s *TaskServer) UpdateTask(ctx context.Context, req *taskv1.UpdateTaskRequest) (*taskv1.UpdateTaskResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid task ID format")
	}

	// Validate input
	if err := grpcerrors.ValidateNotEmpty(req.Title, "title"); err != nil {
		return nil, err
	}
	if err := grpcerrors.ValidateLength(req.Title, "title", grpcerrors.MaxTitleLength); err != nil {
		return nil, err
	}
	if err := grpcerrors.ValidateLength(req.Notes, "notes", grpcerrors.MaxNotesLength); err != nil {
		return nil, err
	}

	task, err := s.service.UpdateTask(ctx, id, req.Title, req.Notes, req.TagNames)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to update task")
	}

	return &taskv1.UpdateTaskResponse{
		Task: taskToProto(task),
	}, nil
}

// DeleteTask deletes a task
func (s *TaskServer) DeleteTask(ctx context.Context, req *taskv1.DeleteTaskRequest) (*taskv1.DeleteTaskResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid task ID format")
	}

	if err := s.service.DeleteTask(ctx, id); err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to delete task")
	}

	return &taskv1.DeleteTaskResponse{}, nil
}

// ListTasks lists tasks with pagination
func (s *TaskServer) ListTasks(ctx context.Context, req *taskv1.ListTasksRequest) (*taskv1.ListTasksResponse, error) {
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

	// Parse filter tag IDs
	filterTagIDs := make([]uuid.UUID, 0, len(req.FilterTagIds))
	for _, tagIDStr := range req.FilterTagIds {
		tagID, err := uuid.Parse(tagIDStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid tag ID format: %s", tagIDStr)
		}
		filterTagIDs = append(filterTagIDs, tagID)
	}

	tasks, err := s.service.ListTasks(ctx, filterTagIDs, pageSize, offset)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to list tasks")
	}

	protoTasks := make([]*taskv1.Task, len(tasks))
	for i, task := range tasks {
		protoTasks[i] = taskToProto(task)
	}

	// Note: next_page_token is not implemented yet
	// Future implementation would return a token when len(tasks) == pageSize
	return &taskv1.ListTasksResponse{
		Tasks: protoTasks,
	}, nil
}

// taskToProto converts a domain Task to a proto Task
func taskToProto(task *domain.Task) *taskv1.Task {
	tagIDs := make([]string, len(task.TagIDs))
	for i, tagID := range task.TagIDs {
		tagIDs[i] = tagID.String()
	}

	return &taskv1.Task{
		Id:        task.ID.String(),
		Title:     task.Title,
		Notes:     task.Notes,
		CreatedAt: timestamppb.New(task.CreatedAt),
		UpdatedAt: timestamppb.New(task.UpdatedAt),
		TagIds:    tagIDs,
	}
}
