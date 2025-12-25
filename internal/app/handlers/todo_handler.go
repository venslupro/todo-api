package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
	todov1 "github.com/venslupro/todo-api/api/gen/todo/v1"
	"github.com/venslupro/todo-api/internal/app/service"
	"github.com/venslupro/todo-api/internal/domain"
	"github.com/venslupro/todo-api/internal/pkg/middleware"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TODOHandler implements the TODOService gRPC interface.
type TODOHandler struct {
	todov1.UnimplementedTODOServiceServer
	service *service.TODOService
}

// NewTODOHandler creates a new TODO handler.
func NewTODOHandler(svc *service.TODOService) *TODOHandler {
	return &TODOHandler{
		service: svc,
	}
}

// CreateTODO creates a new TODO.
func (h *TODOHandler) CreateTODO(ctx context.Context, req *todov1.CreateTODORequest) (*todov1.CreateTODOResponse, error) {
	// Extract user ID from context (set by auth middleware)
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	var status *commonv1.Status
	if req.Status != nil {
		s := req.GetStatus()
		if s != commonv1.Status_STATUS_UNSPECIFIED {
			status = &s
		}
	}

	var priority *commonv1.Priority
	if req.Priority != nil {
		p := req.GetPriority()
		if p != commonv1.Priority_PRIORITY_UNSPECIFIED {
			priority = &p
		}
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		dd := req.DueDate.AsTime()
		dueDate = &dd
	}

	var assignedTo, parentID *string
	if req.AssignedTo != nil {
		assignedTo = req.AssignedTo
	}
	if req.ParentId != nil {
		parentID = req.ParentId
	}

	todo, err := h.service.CreateTODO(ctx, userID, req.Title, description, status, priority, dueDate, req.Tags, assignedTo, parentID)
	if err != nil {
		return nil, err
	}

	return &todov1.CreateTODOResponse{
		Todo: convertToProto(todo),
	}, nil
}

// GetTODO retrieves a TODO by ID.
func (h *TODOHandler) GetTODO(ctx context.Context, req *todov1.GetTODORequest) (*todov1.GetTODOResponse, error) {
	todo, err := h.service.GetTODO(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &todov1.GetTODOResponse{
		Todo: convertToProto(todo),
	}, nil
}

// UpdateTODO updates an existing TODO.
func (h *TODOHandler) UpdateTODO(ctx context.Context, req *todov1.UpdateTODORequest) (*todov1.UpdateTODOResponse, error) {
	var title, description *string
	if req.Title != nil {
		title = req.Title
	}
	if req.Description != nil {
		description = req.Description
	}

	var status *commonv1.Status
	if req.Status != nil {
		s := req.GetStatus()
		if s != commonv1.Status_STATUS_UNSPECIFIED {
			status = &s
		}
	}

	var priority *commonv1.Priority
	if req.Priority != nil {
		p := req.GetPriority()
		if p != commonv1.Priority_PRIORITY_UNSPECIFIED {
			priority = &p
		}
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		dd := req.DueDate.AsTime()
		dueDate = &dd
	}

	var assignedTo, parentID *string
	if req.AssignedTo != nil {
		assignedTo = req.AssignedTo
	}
	if req.ParentId != nil {
		parentID = req.ParentId
	}

	var position *int32
	if req.Position != nil {
		position = req.Position
	}

	todo, err := h.service.UpdateTODO(ctx, req.Id, title, description, status, priority, dueDate, req.Tags, assignedTo, parentID, position)
	if err != nil {
		return nil, err
	}

	return &todov1.UpdateTODOResponse{
		Todo: convertToProto(todo),
	}, nil
}

// DeleteTODO deletes a TODO by ID.
func (h *TODOHandler) DeleteTODO(ctx context.Context, req *todov1.DeleteTODORequest) (*todov1.DeleteTODOResponse, error) {
	if err := h.service.DeleteTODO(ctx, req.Id); err != nil {
		return nil, err
	}

	return &todov1.DeleteTODOResponse{}, nil
}

// ListTODOs retrieves TODOs with filtering, sorting, and pagination.
func (h *TODOHandler) ListTODOs(ctx context.Context, req *todov1.ListTODOsRequest) (*todov1.ListTODOsResponse, error) {
	// Extract user ID from context (set by auth middleware).
	// If authentication fails, return error to ensure security.
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Build sort options from request.
	sortOptions := make([]domain.SortOption, 0, len(req.SortOptions))
	for _, so := range req.SortOptions {
		if so == nil {
			continue
		}
		// Convert proto SortOption to domain SortOption.
		// Note: proto generated code uses v1.SortOption from gen/common/v1,
		// but the actual implementation is in api/gen/common/v1 as commonv1.SortOption.
		// Since the fields are exported, we can access them directly using reflection
		// or convert through commonv1.SortOption.
		var field string
		var descending bool
		// Convert to interface{} first, then to commonv1.SortOption
		// This works because both types have the same structure
		if soInterface := interface{}(so); soInterface != nil {
			// Try to convert to commonv1.SortOption using type assertion
			if commonSO, ok := soInterface.(*commonv1.SortOption); ok {
				field = commonSO.GetField()
				descending = commonSO.GetDescending()
			} else {
				// Fallback: use reflection to access fields
				// Since proto fields are exported, we can access them via reflection
				field = getSortOptionFieldByReflection(so)
				descending = getSortOptionDescendingByReflection(so)
			}
		}
		sortOptions = append(sortOptions, domain.SortOption{
			Field:      field,
			Descending: descending,
		})
	}

	// Get pagination parameters with defaults.
	page := int32(1)
	pageSize := int32(20)
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = req.Pagination.Page
		}
		if req.Pagination.PageSize > 0 {
			pageSize = req.Pagination.PageSize
			// Enforce maximum page size.
			if pageSize > 100 {
				pageSize = 100
			}
		}
	}

	// Convert request to domain filter.
	domainFilter := convertFilter(req, userID)
	todos, pagination, err := h.service.ListTODOs(ctx, domainFilter, sortOptions, page, pageSize)
	if err != nil {
		return nil, err
	}

	// Convert domain TODOs to proto format.
	protoTodos := make([]*todov1.TODO, 0, len(todos))
	for _, todo := range todos {
		protoTodos = append(protoTodos, convertToProto(todo))
	}

	// Build pagination response.
	paginationResp := &commonv1.PaginationResponse{
		TotalItems:  pagination.TotalItems,
		TotalPages:  pagination.TotalPages,
		CurrentPage: pagination.CurrentPage,
		PageSize:    pagination.PageSize,
		HasNext:     pagination.HasNext,
		HasPrev:     pagination.HasPrev,
	}

	return &todov1.ListTODOsResponse{
		Todos:      protoTodos,
		Pagination: paginationResp,
	}, nil
}

// BulkUpdateStatus updates status for multiple TODOs.
func (h *TODOHandler) BulkUpdateStatus(ctx context.Context, req *todov1.BulkUpdateStatusRequest) (*todov1.BulkUpdateStatusResponse, error) {
	status := req.GetStatus()
	if err := h.service.BulkUpdateStatus(ctx, req.Ids, status); err != nil {
		return nil, err
	}

	return &todov1.BulkUpdateStatusResponse{}, nil
}

// BulkDelete deletes multiple TODOs.
func (h *TODOHandler) BulkDelete(ctx context.Context, req *todov1.BulkDeleteRequest) (*todov1.BulkDeleteResponse, error) {
	if err := h.service.BulkDelete(ctx, req.Ids); err != nil {
		return nil, err
	}

	return &todov1.BulkDeleteResponse{}, nil
}

// MoveTODO moves a TODO to a new position or parent.
func (h *TODOHandler) MoveTODO(ctx context.Context, req *todov1.MoveTODORequest) (*todov1.MoveTODOResponse, error) {
	var parentID *string
	if req.ParentId != nil {
		parentID = req.ParentId
	}

	var position *int32
	if req.Position != nil {
		position = req.Position
	}

	todo, err := h.service.MoveTODO(ctx, req.Id, parentID, position)
	if err != nil {
		return nil, err
	}

	return &todov1.MoveTODOResponse{
		Todo: convertToProto(todo),
	}, nil
}

// CompleteTODO marks a TODO as completed.
func (h *TODOHandler) CompleteTODO(ctx context.Context, req *todov1.CompleteTODORequest) (*todov1.CompleteTODOResponse, error) {
	todo, err := h.service.CompleteTODO(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &todov1.CompleteTODOResponse{
		Todo: convertToProto(todo),
	}, nil
}

// ReopenTODO reopens a completed TODO.
func (h *TODOHandler) ReopenTODO(ctx context.Context, req *todov1.ReopenTODORequest) (*todov1.ReopenTODOResponse, error) {
	todo, err := h.service.ReopenTODO(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &todov1.ReopenTODOResponse{
		Todo: convertToProto(todo),
	}, nil
}

// Helper functions

// getSortOptionFieldByReflection extracts the Field value from a proto SortOption using reflection.
func getSortOptionFieldByReflection(so interface{}) string {
	// Use reflection to access the Field field
	// This is a workaround for proto import path mismatch
	v := reflect.ValueOf(so)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		field := v.FieldByName("Field")
		if field.IsValid() && field.Kind() == reflect.String {
			return field.String()
		}
	}
	return ""
}

// getSortOptionDescendingByReflection extracts the Descending value from a proto SortOption using reflection.
func getSortOptionDescendingByReflection(so interface{}) bool {
	// Use reflection to access the Descending field
	// This is a workaround for proto import path mismatch
	v := reflect.ValueOf(so)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		field := v.FieldByName("Descending")
		if field.IsValid() && field.Kind() == reflect.Bool {
			return field.Bool()
		}
	}
	return false
}

// convertToProto converts a domain TODO to a proto TODO message.
func convertToProto(todo *domain.TODO) *todov1.TODO {
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
	if !todo.CreatedAt.IsZero() {
		pb.CreatedAt = timestamppb.New(todo.CreatedAt)
	}
	if !todo.UpdatedAt.IsZero() {
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

	// Note: IsShared and SharedBy are not in the proto definition.
	// If needed, they should be added to the proto file first.

	return pb
}

// convertFilter converts a proto ListTODOsRequest to a domain TODOFilter.
// The userID parameter is the authenticated user's ID, which will be used as the default
// filter unless overridden by req.UserId.
func convertFilter(req *todov1.ListTODOsRequest, userID string) domain.TODOFilter {
	filter := domain.TODOFilter{
		IDs: req.Ids,
	}

	// Use user_id from request if provided, otherwise use authenticated user.
	if req.UserId != nil {
		filter.UserID = req.UserId
	} else {
		filter.UserID = &userID
	}

	// Convert statuses from proto enum to domain enum.
	// The proto uses v1.Status (from gen/common/v1) which is the same underlying type as commonv1.Status.
	// Both are int32-based enums, so we convert through int32.
	if len(req.Statuses) > 0 {
		statuses := make([]commonv1.Status, 0, len(req.Statuses))
		for _, s := range req.Statuses {
			// Convert through int32 since both enum types are based on int32.
			statuses = append(statuses, commonv1.Status(int32(s)))
		}
		filter.Statuses = statuses
	}

	// Convert priorities from proto enum to domain enum.
	// The proto uses v1.Priority (from gen/common/v1) which is the same underlying type as commonv1.Priority.
	// Both are int32-based enums, so we convert through int32.
	if len(req.Priorities) > 0 {
		priorities := make([]commonv1.Priority, 0, len(req.Priorities))
		for _, p := range req.Priorities {
			// Convert through int32 since both enum types are based on int32.
			priorities = append(priorities, commonv1.Priority(int32(p)))
		}
		filter.Priorities = priorities
	}

	// Handle due date range.
	if req.DueDateRange != nil {
		if req.DueDateRange.Start != nil {
			from := req.DueDateRange.Start.AsTime()
			filter.DueDateFrom = &from
		}
		if req.DueDateRange.End != nil {
			to := req.DueDateRange.End.AsTime()
			filter.DueDateTo = &to
		}
	}

	// Copy tags.
	if len(req.Tags) > 0 {
		filter.Tags = req.Tags
	}

	// Handle optional string fields.
	if req.AssignedTo != nil {
		filter.AssignedTo = req.AssignedTo
	}

	if req.ParentId != nil {
		filter.ParentID = req.ParentId
	}

	if req.SearchQuery != nil {
		filter.SearchQuery = req.SearchQuery
		// Enhanced search: search in title, description, and tags by default
		filter.SearchFields = []string{"title", "description", "tags"}
	}

	return filter
}

// HandleAdvancedSearch handles advanced search requests via HTTP query parameters
func (h *TODOHandler) HandleAdvancedSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from context (set by auth middleware)
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	// Build filter from query parameters
	filter := domain.TODOFilter{}

	// User ID (default to authenticated user)
	if userIDParam := query.Get("user_id"); userIDParam != "" {
		filter.UserID = &userIDParam
	} else {
		filter.UserID = &userID
	}

	// Status filter
	if statuses := query["status"]; len(statuses) > 0 {
		for _, statusStr := range statuses {
			if status, ok := commonv1.Status_value[statusStr]; ok {
				filter.Statuses = append(filter.Statuses, commonv1.Status(status))
			}
		}
	}

	// Priority filter
	if priorities := query["priority"]; len(priorities) > 0 {
		for _, priorityStr := range priorities {
			if priority, ok := commonv1.Priority_value[priorityStr]; ok {
				filter.Priorities = append(filter.Priorities, commonv1.Priority(priority))
			}
		}
	}

	// Date range filters
	if createdFrom := query.Get("created_from"); createdFrom != "" {
		if t, err := time.Parse(time.RFC3339, createdFrom); err == nil {
			filter.CreatedDateFrom = &t
		}
	}

	if createdTo := query.Get("created_to"); createdTo != "" {
		if t, err := time.Parse(time.RFC3339, createdTo); err == nil {
			filter.CreatedDateTo = &t
		}
	}

	if completedFrom := query.Get("completed_from"); completedFrom != "" {
		if t, err := time.Parse(time.RFC3339, completedFrom); err == nil {
			filter.CompletedDateFrom = &t
		}
	}

	if completedTo := query.Get("completed_to"); completedTo != "" {
		if t, err := time.Parse(time.RFC3339, completedTo); err == nil {
			filter.CompletedDateTo = &t
		}
	}

	if dueFrom := query.Get("due_from"); dueFrom != "" {
		if t, err := time.Parse(time.RFC3339, dueFrom); err == nil {
			filter.DueDateFrom = &t
		}
	}

	if dueTo := query.Get("due_to"); dueTo != "" {
		if t, err := time.Parse(time.RFC3339, dueTo); err == nil {
			filter.DueDateTo = &t
		}
	}

	// Tags filter
	if tags := query["tag"]; len(tags) > 0 {
		filter.Tags = tags
	}

	// Assigned to filter
	if assignedTo := query.Get("assigned_to"); assignedTo != "" {
		filter.AssignedTo = &assignedTo
	}

	// Parent ID filter
	if parentID := query.Get("parent_id"); parentID != "" {
		filter.ParentID = &parentID
	}

	// Team ID filter
	if teamID := query.Get("team_id"); teamID != "" {
		filter.TeamID = &teamID
	}

	// Shared status filter
	if isShared := query.Get("is_shared"); isShared != "" {
		shared := isShared == "true" || isShared == "1"
		filter.IsShared = &shared
	}

	// Search query
	if searchQuery := query.Get("q"); searchQuery != "" {
		filter.SearchQuery = &searchQuery

		// Search fields
		if searchFields := query["search_field"]; len(searchFields) > 0 {
			filter.SearchFields = searchFields
		} else {
			// Default search fields
			filter.SearchFields = []string{"title", "description", "tags"}
		}
	}

	// Pagination
	page := int32(1)
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := parseInt32(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := int32(20)
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		if ps, err := parseInt32(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
			if pageSize > 100 {
				pageSize = 100
			}
		}
	}

	// Sort options
	var sortOptions []domain.SortOption
	if sortBy := query.Get("sort_by"); sortBy != "" {
		descending := query.Get("sort_order") == "desc"
		sortOptions = append(sortOptions, domain.SortOption{
			Field:      sortBy,
			Descending: descending,
		})
	}

	// Execute search
	todos, pagination, err := h.service.ListTODOs(r.Context(), filter, sortOptions, page, pageSize)
	if err != nil {
		http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to JSON response
	response := map[string]interface{}{
		"todos": convertTODOsToMap(todos),
		"pagination": map[string]interface{}{
			"total_items":  pagination.TotalItems,
			"total_pages":  pagination.TotalPages,
			"current_page": pagination.CurrentPage,
			"page_size":    pagination.PageSize,
			"has_next":     pagination.HasNext,
			"has_prev":     pagination.HasPrev,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper function to parse int32 from string
func parseInt32(s string) (int32, error) {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(i), nil
}

// Helper function to convert TODOs to map for JSON serialization
func convertTODOsToMap(todos []*domain.TODO) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(todos))
	for _, todo := range todos {
		todoMap := map[string]interface{}{
			"id":          todo.ID,
			"user_id":     todo.UserID,
			"title":       todo.Title,
			"description": todo.Description,
			"status":      todo.Status.String(),
			"priority":    todo.Priority.String(),
			"tags":        todo.Tags,
			"position":    todo.Position,
			"created_at":  todo.CreatedAt.Format(time.RFC3339),
			"updated_at":  todo.UpdatedAt.Format(time.RFC3339),
		}

		if todo.DueDate != nil {
			todoMap["due_date"] = todo.DueDate.Format(time.RFC3339)
		}
		if todo.CompletedAt != nil {
			todoMap["completed_at"] = todo.CompletedAt.Format(time.RFC3339)
		}
		if todo.AssignedTo != nil {
			todoMap["assigned_to"] = *todo.AssignedTo
		}
		if todo.ParentID != nil {
			todoMap["parent_id"] = *todo.ParentID
		}
		if todo.TeamID != nil {
			todoMap["team_id"] = *todo.TeamID
		}
		todoMap["is_shared"] = todo.IsShared

		result = append(result, todoMap)
	}
	return result
}
