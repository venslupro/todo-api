package service

import (
	"context"
	"fmt"
	"time"

	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
	todov1 "github.com/venslupro/todo-api/api/gen/todo/v1"
	"github.com/venslupro/todo-api/internal/domain"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TODOService provides business logic for TODO operations
type TODOService struct {
	repo             domain.TODORepository
	websocketService *WebSocketService
}

// NewTODOService creates a new TODO service
func NewTODOService(repo domain.TODORepository, websocketService *WebSocketService) *TODOService {
	return &TODOService{
		repo:             repo,
		websocketService: websocketService,
	}
}

// CreateTODO creates a new TODO
func (s *TODOService) CreateTODO(ctx context.Context, userID, title string, description *string, status *commonv1.Status, priority *commonv1.Priority, dueDate *time.Time, tags []string, assignedTo, parentID *string) (*domain.TODO, error) {
	// Validate title
	if title == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "title is required")
	}

	// Validate user ID
	if userID == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "user_id is required")
	}

	// Create TODO
	todo := domain.NewTODO(userID, title)
	if description != nil {
		todo.Description = *description
	}
	if status != nil {
		todo.Status = *status
	}
	if priority != nil {
		todo.Priority = *priority
	}
	if dueDate != nil {
		todo.DueDate = dueDate
	}
	if tags != nil {
		todo.Tags = tags
	}
	if assignedTo != nil {
		todo.AssignedTo = assignedTo
	}
	if parentID != nil {
		// Validate parent exists
		parentExists, err := s.repo.Exists(ctx, *parentID)
		if err != nil {
			return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to validate parent: %v", err))
		}
		if !parentExists {
			return nil, grpcstatus.Error(codes.NotFound, "parent todo not found")
		}
		todo.ParentID = parentID
	}
	// Save TODO
	if err := s.repo.Create(ctx, todo); err != nil {
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to create TODO: %v", err))
	}

	// Broadcast WebSocket notification
	if s.websocketService != nil {
		s.websocketService.BroadcastTODOUpdate(ctx, todo, "created")
	}

	return todo, nil
}

// GetTODO retrieves a TODO by ID
func (s *TODOService) GetTODO(ctx context.Context, id string) (*domain.TODO, error) {
	if id == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "id is required")
	}

	todo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, grpcstatus.Error(codes.NotFound, fmt.Sprintf("todo not found: %v", err))
	}

	return todo, nil
}

// UpdateTODO updates an existing TODO
func (s *TODOService) UpdateTODO(ctx context.Context, id string, title, description *string, status *commonv1.Status, priority *commonv1.Priority, dueDate *time.Time, tags []string, assignedTo, parentID *string, position *int32) (*domain.TODO, error) {
	if id == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "id is required")
	}

	// Get existing TODO
	todo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, grpcstatus.Error(codes.NotFound, fmt.Sprintf("todo not found: %v", err))
	}

	// Validate parent if provided
	if parentID != nil && *parentID != "" {
		if *parentID == id {
			return nil, grpcstatus.Error(codes.InvalidArgument, "todo cannot be its own parent")
		}
		parentExists, err := s.repo.Exists(ctx, *parentID)
		if err != nil {
			return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to validate parent: %v", err))
		}
		if !parentExists {
			return nil, grpcstatus.Error(codes.NotFound, "parent todo not found")
		}
	}

	// Update TODO
	todo.Update(title, description, status, priority, dueDate, tags, assignedTo, parentID, position)

	if err := s.repo.Update(ctx, todo); err != nil {
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to update todo: %v", err))
	}

	// Broadcast WebSocket notification
	if s.websocketService != nil {
		s.websocketService.BroadcastTODOUpdate(ctx, todo, "updated")
	}

	return todo, nil
}

// DeleteTODO deletes a TODO by ID
func (s *TODOService) DeleteTODO(ctx context.Context, id string) error {
	if id == "" {
		return grpcstatus.Error(codes.InvalidArgument, "id is required")
	}

	// Get TODO before deletion for WebSocket notification
	todo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return grpcstatus.Error(codes.NotFound, fmt.Sprintf("todo not found: %v", err))
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to delete todo: %v", err))
	}

	// Broadcast WebSocket notification
	if s.websocketService != nil {
		s.websocketService.BroadcastTODOUpdate(ctx, todo, "deleted")
	}

	return nil
}

// ListTODOs retrieves TODOs with filtering, sorting, and pagination
func (s *TODOService) ListTODOs(ctx context.Context, filter domain.TODOFilter, sortOptions []domain.SortOption, page, pageSize int32) ([]*domain.TODO, *domain.PaginationResult, error) {
	options := domain.TODOListOptions{
		Filter:      filter,
		SortOptions: sortOptions,
		Page:        page,
		PageSize:    pageSize,
	}

	todos, pagination, err := s.repo.List(ctx, options)
	if err != nil {
		return nil, nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to list todos: %v", err))
	}

	return todos, pagination, nil
}

// BulkUpdateStatus updates status for multiple TODOs
func (s *TODOService) BulkUpdateStatus(ctx context.Context, ids []string, status commonv1.Status) error {
	if len(ids) == 0 {
		return grpcstatus.Error(codes.InvalidArgument, "ids are required")
	}

	if err := s.repo.BulkUpdateStatus(ctx, ids, status); err != nil {
		return grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to bulk update status: %v", err))
	}

	return nil
}

// BulkDelete deletes multiple TODOs
func (s *TODOService) BulkDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return grpcstatus.Error(codes.InvalidArgument, "ids are required")
	}

	if err := s.repo.BulkDelete(ctx, ids); err != nil {
		return grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to bulk delete: %v", err))
	}

	return nil
}

// CompleteTODO marks a TODO as completed
func (s *TODOService) CompleteTODO(ctx context.Context, id string) (*domain.TODO, error) {
	if id == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "id is required")
	}

	todo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, grpcstatus.Error(codes.NotFound, fmt.Sprintf("todo not found: %v", err))
	}

	todo.Complete()

	if err := s.repo.Update(ctx, todo); err != nil {
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to complete todo: %v", err))
	}

	return todo, nil
}

// ReopenTODO reopens a completed TODO
func (s *TODOService) ReopenTODO(ctx context.Context, id string) (*domain.TODO, error) {
	if id == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "id is required")
	}

	todo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, grpcstatus.Error(codes.NotFound, fmt.Sprintf("todo not found: %v", err))
	}

	if !todo.IsCompleted() {
		return nil, grpcstatus.Error(codes.FailedPrecondition, "todo is not completed")
	}

	todo.Reopen()

	if err := s.repo.Update(ctx, todo); err != nil {
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to reopen todo: %v", err))
	}

	return todo, nil
}

// MoveTODO moves a TODO to a new position or parent
func (s *TODOService) MoveTODO(ctx context.Context, id string, parentID *string, position *int32) (*domain.TODO, error) {
	if id == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "id is required")
	}

	todo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, grpcstatus.Error(codes.NotFound, fmt.Sprintf("todo not found: %v", err))
	}

	if parentID != nil && *parentID != "" {
		if *parentID == id {
			return nil, grpcstatus.Error(codes.InvalidArgument, "todo cannot be its own parent")
		}
		parentExists, err := s.repo.Exists(ctx, *parentID)
		if err != nil {
			return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to validate parent: %v", err))
		}
		if !parentExists {
			return nil, grpcstatus.Error(codes.NotFound, "parent todo not found")
		}
	}

	todo.Update(nil, nil, nil, nil, nil, nil, nil, parentID, position)

	if err := s.repo.Update(ctx, todo); err != nil {
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to move todo: %v", err))
	}

	return todo, nil
}

// ConvertToProto converts a domain TODO to proto TODO
func ConvertToProto(todo *domain.TODO) *todov1.TODO {
	pb := &todov1.TODO{
		Id:          todo.ID,
		UserId:      todo.UserID,
		Title:       todo.Title,
		Description: todo.Description,
		Status:      todo.Status,
		Priority:    todo.Priority,
		Tags:        todo.Tags,
		Position:    todo.Position,
	}

	if todo.DueDate != nil {
		pb.DueDate = timestamppb.New(*todo.DueDate)
	}
	if todo.CreatedAt != (time.Time{}) {
		pb.CreatedAt = timestamppb.New(todo.CreatedAt)
	}
	if todo.UpdatedAt != (time.Time{}) {
		pb.UpdatedAt = timestamppb.New(todo.UpdatedAt)
	}
	if todo.CompletedAt != nil {
		pb.CompletedAt = timestamppb.New(*todo.CompletedAt)
	}
	if todo.AssignedTo != nil {
		pb.AssignedTo = *todo.AssignedTo
	}
	if todo.ParentID != nil {
		pb.ParentId = *todo.ParentID
	}

	return pb
}
