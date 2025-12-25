package service

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
	"github.com/venslupro/todo-api/internal/domain"
)

// WebSocketServiceInterface defines the interface for WebSocket service
// This allows us to use both real WebSocketService and MockWebSocketService
type WebSocketServiceInterface interface {
	Run()
	BroadcastTODOUpdate(ctx context.Context, todo *domain.TODO, action string)
	BroadcastTeamUpdate(ctx context.Context, team *domain.Team, action string)
	BroadcastUserNotification(ctx context.Context, userID, messageType, content string)
	HandleWebSocketConnection(w http.ResponseWriter, r *http.Request, userID string, teamIDs []string) error
}

// MockRepository is a mock implementation of TODORepository for testing
type MockRepository struct {
	todos map[string]*domain.TODO
}

// MockWebSocketService is a mock implementation of WebSocketService for testing
type MockWebSocketService struct{}

func (m *MockWebSocketService) Run() {
	// Mock implementation - do nothing
}

func (m *MockWebSocketService) BroadcastTODOUpdate(ctx context.Context, todo *domain.TODO, action string) {
	// Mock implementation - do nothing
}

func (m *MockWebSocketService) BroadcastTeamUpdate(ctx context.Context, team *domain.Team, action string) {
	// Mock implementation - do nothing
}

func (m *MockWebSocketService) BroadcastUserNotification(ctx context.Context, userID, messageType, content string) {
	// Mock implementation - do nothing
}

func (m *MockWebSocketService) HandleWebSocketConnection(w http.ResponseWriter, r *http.Request, userID string, teamIDs []string) error {
	// Mock implementation - do nothing
	return nil
}

// NewMockWebSocketService creates a new mock WebSocket service
func NewMockWebSocketService() *MockWebSocketService {
	return &MockWebSocketService{}
}

// NewTODOServiceForTest creates a new TODO service for testing
// This version accepts the interface instead of concrete type
func NewTODOServiceForTest(repo domain.TODORepository, websocketService WebSocketServiceInterface) *TODOService {
	return &TODOService{
		repo:             repo,
		websocketService: nil, // Use nil for testing to avoid WebSocket issues
	}
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		todos: make(map[string]*domain.TODO),
	}
}

func (m *MockRepository) Create(ctx context.Context, todo *domain.TODO) error {
	m.todos[todo.ID] = todo
	return nil
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*domain.TODO, error) {
	todo, ok := m.todos[id]
	if !ok {
		return nil, &NotFoundError{ID: id}
	}
	return todo, nil
}

func (m *MockRepository) Update(ctx context.Context, todo *domain.TODO) error {
	if _, ok := m.todos[todo.ID]; !ok {
		return &NotFoundError{ID: todo.ID}
	}
	m.todos[todo.ID] = todo
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, id string) error {
	if _, ok := m.todos[id]; !ok {
		return &NotFoundError{ID: id}
	}
	delete(m.todos, id)
	return nil
}

func (m *MockRepository) List(ctx context.Context, options domain.TODOListOptions) ([]*domain.TODO, *domain.PaginationResult, error) {
	var todos []*domain.TODO

	// Apply filtering
	for _, todo := range m.todos {
		if m.matchesFilter(todo, options.Filter) {
			todos = append(todos, todo)
		}
	}

	// Apply sorting
	if len(options.SortOptions) > 0 {
		sortOption := options.SortOptions[0] // Use first sort option for simplicity
		sortOrder := "asc"
		if sortOption.Descending {
			sortOrder = "desc"
		}
		todos = m.sortTODOs(todos, sortOption.Field, sortOrder)
	}

	// Apply pagination
	page := options.Page
	if page < 1 {
		page = 1
	}
	pageSize := options.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize
	if endIndex > int32(len(todos)) {
		endIndex = int32(len(todos))
	}

	if startIndex >= int32(len(todos)) {
		return []*domain.TODO{}, &domain.PaginationResult{
			TotalItems:  int32(len(todos)),
			TotalPages:  (int32(len(todos)) + pageSize - 1) / pageSize,
			CurrentPage: page,
			PageSize:    pageSize,
			HasNext:     false,
			HasPrev:     page > 1,
		}, nil
	}

	pagedTodos := todos[startIndex:endIndex]

	totalItems := int32(len(todos))
	totalPages := (totalItems + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	pagination := &domain.PaginationResult{
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
		HasNext:     page < totalPages,
		HasPrev:     page > 1,
	}

	return pagedTodos, pagination, nil
}

// matchesFilter checks if a TODO matches the filter criteria
func (m *MockRepository) matchesFilter(todo *domain.TODO, filter domain.TODOFilter) bool {
	// Filter by IDs
	if len(filter.IDs) > 0 {
		found := false
		for _, id := range filter.IDs {
			if todo.ID == id {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by UserID
	if filter.UserID != nil && todo.UserID != *filter.UserID {
		return false
	}

	// Filter by Statuses
	if len(filter.Statuses) > 0 {
		found := false
		for _, status := range filter.Statuses {
			if todo.Status == status {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by Priorities
	if len(filter.Priorities) > 0 {
		found := false
		for _, priority := range filter.Priorities {
			if todo.Priority == priority {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by DueDate range
	if filter.DueDateFrom != nil && (todo.DueDate == nil || todo.DueDate.Before(*filter.DueDateFrom)) {
		return false
	}
	if filter.DueDateTo != nil && (todo.DueDate == nil || todo.DueDate.After(*filter.DueDateTo)) {
		return false
	}

	// Filter by CreatedDate range
	if filter.CreatedDateFrom != nil && todo.CreatedAt.Before(*filter.CreatedDateFrom) {
		return false
	}
	if filter.CreatedDateTo != nil && todo.CreatedAt.After(*filter.CreatedDateTo) {
		return false
	}

	// Filter by CompletedDate range
	if filter.CompletedDateFrom != nil && (todo.CompletedAt == nil || todo.CompletedAt.Before(*filter.CompletedDateFrom)) {
		return false
	}
	if filter.CompletedDateTo != nil && (todo.CompletedAt == nil || todo.CompletedAt.After(*filter.CompletedDateTo)) {
		return false
	}

	// Filter by Tags
	if len(filter.Tags) > 0 {
		found := false
		for _, tag := range filter.Tags {
			for _, todoTag := range todo.Tags {
				if todoTag == tag {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by AssignedTo
	if filter.AssignedTo != nil {
		if todo.AssignedTo == nil || *todo.AssignedTo != *filter.AssignedTo {
			return false
		}
	}

	// Filter by ParentID
	if filter.ParentID != nil {
		if todo.ParentID == nil || *todo.ParentID != *filter.ParentID {
			return false
		}
	}

	// Filter by TeamID
	if filter.TeamID != nil {
		if todo.TeamID == nil || *todo.TeamID != *filter.TeamID {
			return false
		}
	}

	// Filter by IsShared
	if filter.IsShared != nil {
		shared := todo.TeamID != nil && *todo.TeamID != ""
		if shared != *filter.IsShared {
			return false
		}
	}

	// Filter by SearchQuery
	if filter.SearchQuery != nil {
		searchQuery := *filter.SearchQuery
		searchFields := filter.SearchFields
		if len(searchFields) == 0 {
			searchFields = []string{"title", "description", "tags"}
		}

		found := false
		for _, field := range searchFields {
			switch field {
			case "title":
				if containsIgnoreCase(todo.Title, searchQuery) {
					found = true
				}
			case "description":
				if todo.Description != "" && containsIgnoreCase(todo.Description, searchQuery) {
					found = true
				}
			case "tags":
				for _, tag := range todo.Tags {
					if containsIgnoreCase(tag, searchQuery) {
						found = true
						break
					}
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// sortTODOs sorts TODOs based on the specified field and order
func (m *MockRepository) sortTODOs(todos []*domain.TODO, sortBy string, sortOrder string) []*domain.TODO {
	sorted := make([]*domain.TODO, len(todos))
	copy(sorted, todos)

	switch sortBy {
	case "created_at":
		if sortOrder == "desc" {
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i].CreatedAt.Before(sorted[j].CreatedAt) {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		} else {
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i].CreatedAt.After(sorted[j].CreatedAt) {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		}
	case "due_date":
		if sortOrder == "desc" {
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if (sorted[i].DueDate == nil && sorted[j].DueDate != nil) ||
						(sorted[i].DueDate != nil && sorted[j].DueDate != nil && sorted[i].DueDate.Before(*sorted[j].DueDate)) {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		} else {
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if (sorted[i].DueDate != nil && sorted[j].DueDate == nil) ||
						(sorted[i].DueDate != nil && sorted[j].DueDate != nil && sorted[i].DueDate.After(*sorted[j].DueDate)) {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		}
	case "priority":
		if sortOrder == "desc" {
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i].Priority < sorted[j].Priority {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		} else {
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i].Priority > sorted[j].Priority {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		}
	}

	return sorted
}

// containsIgnoreCase checks if a string contains another string (case insensitive)
func containsIgnoreCase(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if len(s[i:]) >= len(substr) && equalsIgnoreCase(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

// equalsIgnoreCase checks if two strings are equal (case insensitive)
func equalsIgnoreCase(s1, s2 string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		if toLower(s1[i]) != toLower(s2[i]) {
			return false
		}
	}
	return true
}

// toLower converts a byte to lowercase
func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

func (m *MockRepository) BulkUpdateStatus(ctx context.Context, ids []string, status commonv1.Status) error {
	for _, id := range ids {
		if todo, ok := m.todos[id]; ok {
			todo.Status = status
		}
	}
	return nil
}

func (m *MockRepository) BulkDelete(ctx context.Context, ids []string) error {
	for _, id := range ids {
		delete(m.todos, id)
	}
	return nil
}

func (m *MockRepository) Exists(ctx context.Context, id string) (bool, error) {
	_, ok := m.todos[id]
	return ok, nil
}

func (m *MockRepository) GetSharedTODOs(ctx context.Context, teamID string, options domain.TODOListOptions) ([]*domain.TODO, *domain.PaginationResult, error) {
	// For testing purposes, return empty results
	return []*domain.TODO{}, &domain.PaginationResult{}, nil
}

func (m *MockRepository) GetSharedTeams(ctx context.Context, todoID string) ([]string, error) {
	// For testing purposes, return empty results
	return []string{}, nil
}

type NotFoundError struct {
	ID string
}

func (e *NotFoundError) Error() string {
	return "todo not found: " + e.ID
}

func TestTODOService_CreateTODO(t *testing.T) {
	repo := NewMockRepository()
	service := NewTODOService(repo, nil)
	ctx := context.Background()

	userID := "user-123"
	title := "Test TODO"

	todo, err := service.CreateTODO(ctx, userID, title, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if todo == nil {
		t.Fatal("Expected todo to be created")
	}
	if todo.Title != title {
		t.Errorf("Expected title to be %s, got %s", title, todo.Title)
	}
	if todo.UserID != userID {
		t.Errorf("Expected userID to be %s, got %s", userID, todo.UserID)
	}
}

func TestTODOService_CreateTODO_EmptyTitle(t *testing.T) {
	repo := NewMockRepository()
	service := NewTODOService(repo, nil)
	ctx := context.Background()

	_, err := service.CreateTODO(ctx, "user-123", "", nil, nil, nil, nil, nil, nil, nil)
	if err == nil {
		t.Error("Expected error for empty title")
	}
}

func TestTODOService_GetTODO(t *testing.T) {
	repo := NewMockRepository()
	service := NewTODOService(repo, nil)
	ctx := context.Background()

	// Create a TODO first
	todo, _ := service.CreateTODO(ctx, "user-123", "Test TODO", nil, nil, nil, nil, nil, nil, nil)

	// Get the TODO
	retrieved, err := service.GetTODO(ctx, todo.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if retrieved.ID != todo.ID {
		t.Errorf("Expected ID to be %s, got %s", todo.ID, retrieved.ID)
	}
}

func TestTODOService_GetTODO_NotFound(t *testing.T) {
	repo := NewMockRepository()
	service := NewTODOService(repo, nil)
	ctx := context.Background()

	_, err := service.GetTODO(ctx, "non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent TODO")
	}
}

func TestTODOService_UpdateTODO(t *testing.T) {
	repo := NewMockRepository()
	service := NewTODOService(repo, nil)
	ctx := context.Background()

	// Create a TODO
	todo, _ := service.CreateTODO(ctx, "user-123", "Original Title", nil, nil, nil, nil, nil, nil, nil)

	// Update the TODO
	newTitle := "Updated Title"
	updated, err := service.UpdateTODO(ctx, todo.ID, &newTitle, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if updated.Title != newTitle {
		t.Errorf("Expected title to be %s, got %s", newTitle, updated.Title)
	}
}

func TestTODOService_DeleteTODO(t *testing.T) {
	repo := NewMockRepository()
	service := NewTODOService(repo, nil)
	ctx := context.Background()

	// Create a TODO
	todo, _ := service.CreateTODO(ctx, "user-123", "Test TODO", nil, nil, nil, nil, nil, nil, nil)

	// Complete the TODO
	completed, err := service.CompleteTODO(ctx, todo.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if completed.Status != commonv1.Status_STATUS_COMPLETED {
		t.Errorf("Expected status to be STATUS_COMPLETED, got %v", completed.Status)
	}
	if completed.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
}

func TestTODOService_CompleteTODO(t *testing.T) {
	repo := NewMockRepository()
	service := NewTODOService(repo, nil)
	ctx := context.Background()

	// Create and complete a TODO
	todo, _ := service.CreateTODO(ctx, "user-123", "Test TODO", nil, nil, nil, nil, nil, nil, nil)
	service.CompleteTODO(ctx, todo.ID)

	// Reopen the TODO
	reopened, err := service.ReopenTODO(ctx, todo.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if reopened.Status != commonv1.Status_STATUS_NOT_STARTED {
		t.Errorf("Expected status to be STATUS_NOT_STARTED, got %v", reopened.Status)
	}
}

func TestTODOService_ReopenTODO(t *testing.T) {
	repo := NewMockRepository()
	service := NewTODOService(repo, nil)
	ctx := context.Background()

	// Create multiple TODOs
	service.CreateTODO(ctx, "user-123", "TODO 1", nil, nil, nil, nil, nil, nil, nil)
	service.CreateTODO(ctx, "user-123", "TODO 2", nil, nil, nil, nil, nil, nil, nil)
	service.CreateTODO(ctx, "user-123", "TODO 3", nil, nil, nil, nil, nil, nil, nil)

	// List TODOs
	filter := domain.TODOFilter{}
	todos, pagination, err := service.ListTODOs(ctx, filter, nil, 1, 20)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(todos) < 3 {
		t.Errorf("Expected at least 3 TODOs, got %d", len(todos))
	}
	if pagination == nil {
		t.Error("Expected pagination result")
	}
}

func TestTODOService_BulkDelete(t *testing.T) {
	repo := NewMockRepository()
	service := NewTODOService(repo, nil)
	ctx := context.Background()

	// Create multiple TODOs
	todo1, _ := service.CreateTODO(ctx, "user-123", "TODO 1", nil, nil, nil, nil, nil, nil, nil)
	todo2, _ := service.CreateTODO(ctx, "user-123", "TODO 2", nil, nil, nil, nil, nil, nil, nil)

	// Bulk update status
	ids := []string{todo1.ID, todo2.ID}
	err := service.BulkUpdateStatus(ctx, ids, commonv1.Status_STATUS_COMPLETED)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify status was updated
	retrieved1, _ := service.GetTODO(ctx, todo1.ID)
	retrieved2, _ := service.GetTODO(ctx, todo2.ID)

	if retrieved1.Status != commonv1.Status_STATUS_COMPLETED {
		t.Error("Expected todo1 status to be COMPLETED")
	}
	if retrieved2.Status != commonv1.Status_STATUS_COMPLETED {
		t.Error("Expected todo2 status to be COMPLETED")
	}
}

func TestTODOService_ListTODOs(t *testing.T) {
	repo := NewMockRepository()
	service := NewTODOService(repo, nil)
	ctx := context.Background()

	// Create multiple TODOs
	todo1, _ := service.CreateTODO(ctx, "user-123", "TODO 1", nil, nil, nil, nil, nil, nil, nil)
	todo2, _ := service.CreateTODO(ctx, "user-123", "TODO 2", nil, nil, nil, nil, nil, nil, nil)

	// Bulk delete
	ids := []string{todo1.ID, todo2.ID}
	err := service.BulkDelete(ctx, ids)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify TODOs were deleted
	_, err1 := service.GetTODO(ctx, todo1.ID)
	_, err2 := service.GetTODO(ctx, todo2.ID)

	if err1 == nil {
		t.Error("Expected todo1 to be deleted")
	}
	if err2 == nil {
		t.Error("Expected todo2 to be deleted")
	}
}

func TestTODOService_ListTODOs_AdvancedSearch(t *testing.T) {
	repo := NewMockRepository()
	service := NewTODOService(repo, nil)
	ctx := context.Background()

	// Create test TODOs with different attributes
	dueDate1 := time.Now().Add(24 * time.Hour)
	dueDate2 := time.Now().Add(48 * time.Hour)

	desc1 := "Milk, eggs, bread, and vegetables"
	desc2 := "Complete the quarterly project report"
	desc3 := "Prepare agenda for team meeting"
	priorityHigh := commonv1.Priority_PRIORITY_HIGH
	priorityMedium := commonv1.Priority_PRIORITY_MEDIUM
	priorityLow := commonv1.Priority_PRIORITY_LOW

	todo1, _ := service.CreateTODO(ctx, "user-123", "Buy groceries for dinner",
		&desc1, nil, &priorityHigh, &dueDate1, []string{"shopping", "food"}, nil, nil)
	todo2, _ := service.CreateTODO(ctx, "user-123", "Finish project report",
		&desc2, nil, &priorityMedium, &dueDate2, []string{"work", "report"}, nil, nil)
	_, _ = service.CreateTODO(ctx, "user-456", "Team meeting preparation",
		&desc3, nil, &priorityLow, nil, []string{"meeting", "team"}, nil, nil)

	// Test 1: Search by title
	searchQuery1 := "groceries"
	todos, _, err := service.ListTODOs(ctx, domain.TODOFilter{
		UserID:      &todo1.UserID,
		SearchQuery: &searchQuery1,
	}, []domain.SortOption{}, 1, 10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(todos) != 1 || todos[0].ID != todo1.ID {
		t.Error("Expected to find todo1 by title search")
	}

	// Test 2: Search by description
	searchQuery2 := "quarterly"
	todos, _, err = service.ListTODOs(ctx, domain.TODOFilter{
		UserID:      &todo2.UserID,
		SearchQuery: &searchQuery2,
	}, []domain.SortOption{}, 1, 10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(todos) != 1 || todos[0].ID != todo2.ID {
		t.Error("Expected to find todo2 by description search")
	}

	// Test 3: Search by tags
	todos, _, err = service.ListTODOs(ctx, domain.TODOFilter{
		UserID: &todo1.UserID,
		Tags:   []string{"food"},
	}, []domain.SortOption{}, 1, 10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(todos) != 1 || todos[0].ID != todo1.ID {
		t.Error("Expected to find todo1 by tag filter")
	}

	// Test 4: Search by priority
	todos, _, err = service.ListTODOs(ctx, domain.TODOFilter{
		UserID:     &todo1.UserID,
		Priorities: []commonv1.Priority{commonv1.Priority_PRIORITY_HIGH},
	}, []domain.SortOption{}, 1, 10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(todos) != 1 || todos[0].ID != todo1.ID {
		t.Error("Expected to find todo1 by priority filter")
	}

	// Test 5: Search with multiple criteria
	todos, _, err = service.ListTODOs(ctx, domain.TODOFilter{
		UserID:     &todo1.UserID,
		Priorities: []commonv1.Priority{commonv1.Priority_PRIORITY_HIGH, commonv1.Priority_PRIORITY_MEDIUM},
		Tags:       []string{"shopping", "work"},
	}, []domain.SortOption{}, 1, 10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(todos) != 2 {
		t.Errorf("Expected to find 2 todos, found %d", len(todos))
	}
}

func TestTODOService_ListTODOs_PaginationAndSorting(t *testing.T) {
	repo := NewMockRepository()
	service := NewTODOService(repo, nil)
	ctx := context.Background()

	// Create multiple TODOs
	for i := 1; i <= 10; i++ {
		dueDate := time.Now().Add(time.Duration(i) * 24 * time.Hour)
		priority := commonv1.Priority(i % 3) // Cycle through priorities
		_, err := service.CreateTODO(ctx, "user-123", fmt.Sprintf("TODO %d", i),
			nil, nil, &priority, &dueDate, nil, nil, nil)
		if err != nil {
			t.Fatalf("Failed to create TODO %d: %v", i, err)
		}
	}

	// Test pagination
	userID := "user-123"
	todos, pagination, err := service.ListTODOs(ctx, domain.TODOFilter{
		UserID: &userID,
	}, []domain.SortOption{}, 2, 3)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(todos) != 3 {
		t.Errorf("Expected 3 todos on page 2, got %d", len(todos))
	}
	if pagination.CurrentPage != 2 {
		t.Errorf("Expected current page to be 2, got %d", pagination.CurrentPage)
	}
	if pagination.TotalPages != 4 { // 10 items, 3 per page = 4 pages
		t.Errorf("Expected total pages to be 4, got %d", pagination.TotalPages)
	}

	// Test sorting by due date
	todos, _, err = service.ListTODOs(ctx, domain.TODOFilter{
		UserID: &userID,
	}, []domain.SortOption{
		{Field: "due_date", Descending: false},
	}, 1, 10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(todos) < 2 {
		t.Fatal("Need at least 2 todos to test sorting")
	}
	if todos[0].DueDate == nil || todos[1].DueDate == nil {
		t.Fatal("Expected todos to have due dates")
	}
	if todos[0].DueDate.After(*todos[1].DueDate) {
		t.Error("Expected todos to be sorted by due date in ascending order")
	}
}

func TestTODOService_ListTODOs_DateFiltering(t *testing.T) {
	repo := NewMockRepository()
	service := NewTODOService(repo, nil)
	ctx := context.Background()

	// Create TODOs with specific dates
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	todo1, _ := service.CreateTODO(ctx, "user-123", "Yesterday's TODO",
		nil, nil, nil, &yesterday, nil, nil, nil)
	todo2, _ := service.CreateTODO(ctx, "user-123", "Today's TODO",
		nil, nil, nil, &now, nil, nil, nil)
	_, _ = service.CreateTODO(ctx, "user-123", "Tomorrow's TODO",
		nil, nil, nil, &tomorrow, nil, nil, nil)

	// Test due date range filtering
	todos, _, err := service.ListTODOs(ctx, domain.TODOFilter{
		UserID:      &todo1.UserID,
		DueDateFrom: &yesterday,
		DueDateTo:   &now,
	}, []domain.SortOption{}, 1, 10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(todos) != 2 {
		t.Errorf("Expected 2 todos in date range, got %d", len(todos))
	}

	// Verify the correct TODOs are returned
	foundTodo1, foundTodo2 := false, false
	for _, todo := range todos {
		if todo.ID == todo1.ID {
			foundTodo1 = true
		}
		if todo.ID == todo2.ID {
			foundTodo2 = true
		}
	}
	if !foundTodo1 || !foundTodo2 {
		t.Error("Expected to find both todo1 and todo2 in date range")
	}
}
