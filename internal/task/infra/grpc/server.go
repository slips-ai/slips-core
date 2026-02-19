package grpc

import (
	"context"
	"errors"
	"time"

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

	// Parse and validate start_date
	startDate, err := parseStartDateForCreate(req.StartDate)
	if err != nil {
		return nil, err
	}

	task, err := s.service.CreateTask(ctx, req.Title, req.Notes, req.TagNames, startDate)
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

	// Parse and validate start_date only if provided.
	// If field is absent, treat that as "no change" to the task's start date.
	var startDateProvided bool
	var startDate *time.Time
	if req.StartDate != nil {
		startDateProvided = true
		date, err := parseStartDateForUpdate(req.StartDate)
		if err != nil {
			return nil, err
		}
		startDate = date
	}

	task, err := s.service.UpdateTask(ctx, id, req.Title, req.Notes, req.TagNames, startDateProvided, startDate)
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

	// Parse archive filter options
	includeArchived := req.IncludeArchived != nil && *req.IncludeArchived
	archivedOnly := req.ArchivedOnly != nil && *req.ArchivedOnly

	tasks, err := s.service.ListTasks(ctx, filterTagIDs, pageSize, offset, includeArchived, archivedOnly)
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

	checklistItems := make([]*taskv1.ChecklistItem, len(task.Checklist))
	for i := range task.Checklist {
		checklistItems[i] = checklistItemToProto(&task.Checklist[i])
	}

	protoTask := &taskv1.Task{
		Id:             task.ID.String(),
		Title:          task.Title,
		Notes:          task.Notes,
		CreatedAt:      timestamppb.New(task.CreatedAt),
		UpdatedAt:      timestamppb.New(task.UpdatedAt),
		TagIds:         tagIDs,
		ChecklistItems: checklistItems,
	}

	if task.ArchivedAt != nil {
		protoTask.ArchivedAt = timestamppb.New(*task.ArchivedAt)
	}

	if task.StartDate != nil {
		formatted := task.StartDate.Format("2006-01-02")
		protoTask.StartDate = &formatted
	}

	return protoTask
}

func checklistItemToProto(item *domain.ChecklistItem) *taskv1.ChecklistItem {
	return &taskv1.ChecklistItem{
		Id:        item.ID.String(),
		TaskId:    item.TaskID.String(),
		Content:   item.Content,
		Completed: item.Completed,
		SortOrder: item.SortOrder,
		CreatedAt: timestamppb.New(item.CreatedAt),
		UpdatedAt: timestamppb.New(item.UpdatedAt),
	}
}

// parseStartDateForCreate parses and validates optional start_date for create requests.
// nil means inbox.
func parseStartDateForCreate(datePtr *string) (*time.Time, error) {
	if datePtr == nil || *datePtr == "" {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", *datePtr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid start_date format: expected YYYY-MM-DD")
	}

	return &parsed, nil
}

// parseStartDateForUpdate parses and validates optional start_date for update requests.
// empty string clears start_date and moves task to inbox.
func parseStartDateForUpdate(datePtr *string) (*time.Time, error) {
	if datePtr == nil || *datePtr == "" {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", *datePtr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid start_date format: expected YYYY-MM-DD")
	}

	return &parsed, nil
}

// ArchiveTask archives a task
func (s *TaskServer) ArchiveTask(ctx context.Context, req *taskv1.ArchiveTaskRequest) (*taskv1.ArchiveTaskResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid task ID format")
	}

	task, err := s.service.ArchiveTask(ctx, id)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to archive task")
	}

	return &taskv1.ArchiveTaskResponse{
		Task: taskToProto(task),
	}, nil
}

// UnarchiveTask unarchives a task
func (s *TaskServer) UnarchiveTask(ctx context.Context, req *taskv1.UnarchiveTaskRequest) (*taskv1.UnarchiveTaskResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid task ID format")
	}

	task, err := s.service.UnarchiveTask(ctx, id)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to unarchive task")
	}

	return &taskv1.UnarchiveTaskResponse{
		Task: taskToProto(task),
	}, nil
}

// AddChecklistItem creates a checklist item for a task.
func (s *TaskServer) AddChecklistItem(ctx context.Context, req *taskv1.AddChecklistItemRequest) (*taskv1.AddChecklistItemResponse, error) {
	taskID, err := uuid.Parse(req.TaskId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid task ID format")
	}
	if err := grpcerrors.ValidateNotEmpty(req.Content, "content"); err != nil {
		return nil, err
	}
	if err := grpcerrors.ValidateLength(req.Content, "content", grpcerrors.MaxChecklistItemLength); err != nil {
		return nil, err
	}

	item, err := s.service.AddChecklistItem(ctx, taskID, req.Content)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to add checklist item")
	}

	return &taskv1.AddChecklistItemResponse{Item: checklistItemToProto(item)}, nil
}

// UpdateChecklistItem updates checklist item content.
func (s *TaskServer) UpdateChecklistItem(ctx context.Context, req *taskv1.UpdateChecklistItemRequest) (*taskv1.UpdateChecklistItemResponse, error) {
	itemID, err := uuid.Parse(req.ItemId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid checklist item ID format")
	}
	if err := grpcerrors.ValidateNotEmpty(req.Content, "content"); err != nil {
		return nil, err
	}
	if err := grpcerrors.ValidateLength(req.Content, "content", grpcerrors.MaxChecklistItemLength); err != nil {
		return nil, err
	}

	item, err := s.service.UpdateChecklistItemContent(ctx, itemID, req.Content)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to update checklist item")
	}

	return &taskv1.UpdateChecklistItemResponse{Item: checklistItemToProto(item)}, nil
}

// SetChecklistItemCompleted sets checklist completion state.
func (s *TaskServer) SetChecklistItemCompleted(ctx context.Context, req *taskv1.SetChecklistItemCompletedRequest) (*taskv1.SetChecklistItemCompletedResponse, error) {
	itemID, err := uuid.Parse(req.ItemId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid checklist item ID format")
	}

	item, err := s.service.SetChecklistItemCompleted(ctx, itemID, req.Completed)
	if err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to set checklist item completion")
	}

	return &taskv1.SetChecklistItemCompletedResponse{Item: checklistItemToProto(item)}, nil
}

// DeleteChecklistItem deletes checklist item.
func (s *TaskServer) DeleteChecklistItem(ctx context.Context, req *taskv1.DeleteChecklistItemRequest) (*taskv1.DeleteChecklistItemResponse, error) {
	itemID, err := uuid.Parse(req.ItemId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid checklist item ID format")
	}

	if err := s.service.DeleteChecklistItem(ctx, itemID); err != nil {
		return nil, grpcerrors.ToGRPCError(err, "failed to delete checklist item")
	}

	return &taskv1.DeleteChecklistItemResponse{}, nil
}

// ReorderChecklistItems updates checklist ordering for a task.
func (s *TaskServer) ReorderChecklistItems(ctx context.Context, req *taskv1.ReorderChecklistItemsRequest) (*taskv1.ReorderChecklistItemsResponse, error) {
	taskID, err := uuid.Parse(req.TaskId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid task ID format")
	}
	if len(req.ItemIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "item_ids cannot be empty")
	}

	itemIDs := make([]uuid.UUID, len(req.ItemIds))
	for i, itemIDStr := range req.ItemIds {
		itemID, parseErr := uuid.Parse(itemIDStr)
		if parseErr != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid checklist item ID format")
		}
		itemIDs[i] = itemID
	}

	items, err := s.service.ReorderChecklistItems(ctx, taskID, itemIDs)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidChecklistOrder) {
			return nil, status.Error(codes.InvalidArgument, "item_ids must include all checklist item IDs exactly once")
		}
		return nil, grpcerrors.ToGRPCError(err, "failed to reorder checklist items")
	}

	protoItems := make([]*taskv1.ChecklistItem, len(items))
	for i := range items {
		protoItems[i] = checklistItemToProto(&items[i])
	}

	return &taskv1.ReorderChecklistItemsResponse{Items: protoItems}, nil
}
