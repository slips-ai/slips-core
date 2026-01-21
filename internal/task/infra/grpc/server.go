package grpc

import (
	"context"

	"github.com/google/uuid"
	taskv1 "github.com/slips-ai/slips-core/gen/api/proto/task/v1"
	"github.com/slips-ai/slips-core/internal/task/application"
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

	task, err := s.service.CreateTask(ctx, req.Title, req.Notes)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to create task")
	}

	return &taskv1.CreateTaskResponse{
		Task: &taskv1.Task{
			Id:        task.ID.String(),
			Title:     task.Title,
			Notes:     task.Notes,
			CreatedAt: timestamppb.New(task.CreatedAt),
			UpdatedAt: timestamppb.New(task.UpdatedAt),
		},
	}, nil
}

// GetTask retrieves a task by ID
func (s *TaskServer) GetTask(ctx context.Context, req *taskv1.GetTaskRequest) (*taskv1.GetTaskResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid task ID: %v", err)
	}

	task, err := s.service.GetTask(ctx, id)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to get task")
	}

	return &taskv1.GetTaskResponse{
		Task: &taskv1.Task{
			Id:        task.ID.String(),
			Title:     task.Title,
			Notes:     task.Notes,
			CreatedAt: timestamppb.New(task.CreatedAt),
			UpdatedAt: timestamppb.New(task.UpdatedAt),
		},
	}, nil
}

// UpdateTask updates a task
func (s *TaskServer) UpdateTask(ctx context.Context, req *taskv1.UpdateTaskRequest) (*taskv1.UpdateTaskResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid task ID: %v", err)
	}

	// Validate input
	if err := grpcerrors.ValidateNotEmpty(req.Title, "title"); err != nil {
		return nil, err
	}

	task, err := s.service.UpdateTask(ctx, id, req.Title, req.Notes)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to update task")
	}

	return &taskv1.UpdateTaskResponse{
		Task: &taskv1.Task{
			Id:        task.ID.String(),
			Title:     task.Title,
			Notes:     task.Notes,
			CreatedAt: timestamppb.New(task.CreatedAt),
			UpdatedAt: timestamppb.New(task.UpdatedAt),
		},
	}, nil
}

// DeleteTask deletes a task
func (s *TaskServer) DeleteTask(ctx context.Context, req *taskv1.DeleteTaskRequest) (*taskv1.DeleteTaskResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid task ID: %v", err)
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

	tasks, err := s.service.ListTasks(ctx, pageSize, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list tasks: %v", err)
	}

	protoTasks := make([]*taskv1.Task, len(tasks))
	for i, task := range tasks {
		protoTasks[i] = &taskv1.Task{
			Id:        task.ID.String(),
			Title:     task.Title,
			Notes:     task.Notes,
			CreatedAt: timestamppb.New(task.CreatedAt),
			UpdatedAt: timestamppb.New(task.UpdatedAt),
		}
	}

	// Note: next_page_token is not implemented yet
	// Future implementation would return a token when len(tasks) == pageSize
	return &taskv1.ListTasksResponse{
		Tasks: protoTasks,
	}, nil
}
