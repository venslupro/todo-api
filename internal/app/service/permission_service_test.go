package service

import (
	"context"
	"testing"

	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
	"github.com/venslupro/todo-api/internal/domain"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

// MockTODORepository is a mock implementation of TODORepository for testing
type MockTODORepository struct {
	todos       map[string]*domain.TODO
	sharedTODOs map[string][]string // todoID -> teamIDs
}

func NewMockTODORepository() *MockTODORepository {
	return &MockTODORepository{
		todos:       make(map[string]*domain.TODO),
		sharedTODOs: make(map[string][]string),
	}
}

func (m *MockTODORepository) Create(ctx context.Context, todo *domain.TODO) error {
	m.todos[todo.ID] = todo
	return nil
}

func (m *MockTODORepository) GetByID(ctx context.Context, id string) (*domain.TODO, error) {
	todo, ok := m.todos[id]
	if !ok {
		return nil, &NotFoundError{ID: id}
	}
	return todo, nil
}

func (m *MockTODORepository) Update(ctx context.Context, todo *domain.TODO) error {
	if _, ok := m.todos[todo.ID]; !ok {
		return &NotFoundError{ID: todo.ID}
	}
	m.todos[todo.ID] = todo
	return nil
}

func (m *MockTODORepository) Delete(ctx context.Context, id string) error {
	if _, ok := m.todos[id]; !ok {
		return &NotFoundError{ID: id}
	}
	delete(m.todos, id)
	return nil
}

func (m *MockTODORepository) List(ctx context.Context, options domain.TODOListOptions) ([]*domain.TODO, *domain.PaginationResult, error) {
	var todos []*domain.TODO
	for _, todo := range m.todos {
		todos = append(todos, todo)
	}
	total := int32(len(todos))
	pageSize := int32(10)
	if options.PageSize > 0 {
		pageSize = options.PageSize
	}
	totalPages := total / pageSize
	if total%pageSize > 0 {
		totalPages++
	}
	return todos, &domain.PaginationResult{
		TotalItems:  total,
		TotalPages:  totalPages,
		CurrentPage: options.Page,
		PageSize:    pageSize,
		HasNext:     options.Page < totalPages,
		HasPrev:     options.Page > 1,
	}, nil
}

func (m *MockTODORepository) BulkUpdateStatus(ctx context.Context, ids []string, status commonv1.Status) error {
	for _, id := range ids {
		if todo, ok := m.todos[id]; ok {
			todo.Status = status
		}
	}
	return nil
}

func (m *MockTODORepository) BulkDelete(ctx context.Context, ids []string) error {
	for _, id := range ids {
		delete(m.todos, id)
	}
	return nil
}

func (m *MockTODORepository) Exists(ctx context.Context, id string) (bool, error) {
	_, ok := m.todos[id]
	return ok, nil
}

func (m *MockTODORepository) GetSharedTODOs(ctx context.Context, teamID string, options domain.TODOListOptions) ([]*domain.TODO, *domain.PaginationResult, error) {
	var sharedTODOs []*domain.TODO
	for todoID, teams := range m.sharedTODOs {
		for _, team := range teams {
			if team == teamID {
				if todo, ok := m.todos[todoID]; ok {
					sharedTODOs = append(sharedTODOs, todo)
				}
				break
			}
		}
	}
	total := int32(len(sharedTODOs))
	pageSize := int32(10)
	if options.PageSize > 0 {
		pageSize = options.PageSize
	}
	totalPages := total / pageSize
	if total%pageSize > 0 {
		totalPages++
	}
	return sharedTODOs, &domain.PaginationResult{
		TotalItems:  total,
		TotalPages:  totalPages,
		CurrentPage: options.Page,
		PageSize:    pageSize,
		HasNext:     options.Page < totalPages,
		HasPrev:     options.Page > 1,
	}, nil
}

func (m *MockTODORepository) GetSharedTeams(ctx context.Context, todoID string) ([]string, error) {
	teams, ok := m.sharedTODOs[todoID]
	if !ok {
		return []string{}, nil
	}
	return teams, nil
}

func (m *MockTODORepository) ShareTODOWithTeam(ctx context.Context, todoID, teamID string) error {
	if _, ok := m.todos[todoID]; !ok {
		return &NotFoundError{ID: todoID}
	}

	// Check if already shared
	for _, team := range m.sharedTODOs[todoID] {
		if team == teamID {
			return nil // Already shared
		}
	}

	m.sharedTODOs[todoID] = append(m.sharedTODOs[todoID], teamID)
	return nil
}

// MockTeamRepository is a mock implementation of TeamRepository for testing
type MockTeamRepository struct {
	teams       map[string]*domain.Team
	members     map[string]map[string]*domain.TeamMember // teamID -> userID -> member
	sharedTODOs map[string][]string                      // teamID -> todoIDs
}

func NewMockTeamRepository() *MockTeamRepository {
	return &MockTeamRepository{
		teams:       make(map[string]*domain.Team),
		members:     make(map[string]map[string]*domain.TeamMember),
		sharedTODOs: make(map[string][]string),
	}
}

func (m *MockTeamRepository) Create(ctx context.Context, team *domain.Team) error {
	m.teams[team.ID] = team
	return nil
}

func (m *MockTeamRepository) GetByID(ctx context.Context, id string) (*domain.Team, error) {
	team, ok := m.teams[id]
	if !ok {
		return nil, &NotFoundError{ID: id}
	}
	return team, nil
}

func (m *MockTeamRepository) Update(ctx context.Context, team *domain.Team) error {
	if _, ok := m.teams[team.ID]; !ok {
		return &NotFoundError{ID: team.ID}
	}
	m.teams[team.ID] = team
	return nil
}

func (m *MockTeamRepository) Delete(ctx context.Context, id string) error {
	if _, ok := m.teams[id]; !ok {
		return &NotFoundError{ID: id}
	}
	delete(m.teams, id)
	delete(m.members, id)
	delete(m.sharedTODOs, id)
	return nil
}

func (m *MockTeamRepository) ListByUser(ctx context.Context, userID string) ([]*domain.Team, error) {
	var userTeams []*domain.Team
	for teamID, members := range m.members {
		if _, ok := members[userID]; ok {
			if team, ok := m.teams[teamID]; ok {
				userTeams = append(userTeams, team)
			}
		}
	}
	return userTeams, nil
}

func (m *MockTeamRepository) AddMember(ctx context.Context, member *domain.TeamMember) error {
	if _, ok := m.members[member.TeamID]; !ok {
		m.members[member.TeamID] = make(map[string]*domain.TeamMember)
	}
	m.members[member.TeamID][member.UserID] = member
	return nil
}

func (m *MockTeamRepository) RemoveMember(ctx context.Context, teamID, userID string) error {
	if teamMembers, ok := m.members[teamID]; ok {
		delete(teamMembers, userID)
	}
	return nil
}

func (m *MockTeamRepository) GetMember(ctx context.Context, teamID, userID string) (*domain.TeamMember, error) {
	if teamMembers, ok := m.members[teamID]; ok {
		if member, ok := teamMembers[userID]; ok {
			return member, nil
		}
	}
	return nil, &NotFoundError{ID: userID}
}

func (m *MockTeamRepository) ListMembers(ctx context.Context, teamID string) ([]*domain.TeamMember, error) {
	var members []*domain.TeamMember
	if teamMembers, ok := m.members[teamID]; ok {
		for _, member := range teamMembers {
			members = append(members, member)
		}
	}
	return members, nil
}

func (m *MockTeamRepository) UpdateMemberRole(ctx context.Context, teamID, userID string, role commonv1.Role) error {
	if teamMembers, ok := m.members[teamID]; ok {
		if member, ok := teamMembers[userID]; ok {
			member.Role = role
			return nil
		}
	}
	return &NotFoundError{ID: userID}
}

func (m *MockTeamRepository) ShareTODO(ctx context.Context, todoID, teamID, sharedBy string) error {
	if _, ok := m.teams[teamID]; !ok {
		return &NotFoundError{ID: teamID}
	}

	// Check if already shared
	for _, todo := range m.sharedTODOs[teamID] {
		if todo == todoID {
			return nil // Already shared
		}
	}

	m.sharedTODOs[teamID] = append(m.sharedTODOs[teamID], todoID)
	return nil
}

func (m *MockTeamRepository) UnshareTODO(ctx context.Context, todoID, teamID string) error {
	if _, ok := m.teams[teamID]; !ok {
		return &NotFoundError{ID: teamID}
	}

	todos := m.sharedTODOs[teamID]
	for i, todo := range todos {
		if todo == todoID {
			m.sharedTODOs[teamID] = append(todos[:i], todos[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *MockTeamRepository) GetSharedTODOs(ctx context.Context, teamID string) ([]string, error) {
	todos, ok := m.sharedTODOs[teamID]
	if !ok {
		return []string{}, nil
	}
	return todos, nil
}

func (m *MockTeamRepository) GetSharedTeams(ctx context.Context, todoID string) ([]string, error) {
	var teamIDs []string
	for teamID, todos := range m.sharedTODOs {
		for _, todo := range todos {
			if todo == todoID {
				teamIDs = append(teamIDs, teamID)
				break
			}
		}
	}
	return teamIDs, nil
}

func TestPermissionService_CheckTODOPermission(t *testing.T) {
	ctx := context.Background()
	todoRepo := NewMockTODORepository()
	teamRepo := NewMockTeamRepository()
	permissionService := NewPermissionService(todoRepo, teamRepo)

	// Create test data
	ownerID := "owner-user"
	memberID := "member-user"
	adminID := "admin-user"
	nonMemberID := "non-member-user"

	todoID := "test-todo"
	teamID := "test-team"

	// Create a TODO owned by owner
	todo := &domain.TODO{
		ID:     todoID,
		UserID: ownerID,
		Title:  "Test TODO",
	}
	todoRepo.todos[todoID] = todo

	// Create team and add members
	team := &domain.Team{
		ID:        teamID,
		Name:      "Test Team",
		CreatedBy: ownerID,
	}
	teamRepo.teams[teamID] = team

	// Add members to team
	teamRepo.members[teamID] = make(map[string]*domain.TeamMember)
	teamRepo.members[teamID][memberID] = &domain.TeamMember{
		TeamID: teamID,
		UserID: memberID,
		Role:   commonv1.Role_ROLE_MEMBER,
	}
	teamRepo.members[teamID][adminID] = &domain.TeamMember{
		TeamID: teamID,
		UserID: adminID,
		Role:   commonv1.Role_ROLE_ADMIN,
	}

	tests := []struct {
		name           string
		userID         string
		todoID         string
		permission     string
		sharedWithTeam bool
		expectError    bool
		errorCode      codes.Code
	}{
		{
			name:        "owner has view permission",
			userID:      ownerID,
			todoID:      todoID,
			permission:  "view",
			expectError: false,
		},
		{
			name:        "owner has edit permission",
			userID:      ownerID,
			todoID:      todoID,
			permission:  "edit",
			expectError: false,
		},
		{
			name:           "team member has view permission when shared",
			userID:         memberID,
			todoID:         todoID,
			permission:     "view",
			sharedWithTeam: true,
			expectError:    false,
		},
		{
			name:           "team member has edit permission when shared",
			userID:         memberID,
			todoID:         todoID,
			permission:     "edit",
			sharedWithTeam: true,
			expectError:    false,
		},
		{
			name:           "team admin has admin permission when shared",
			userID:         adminID,
			todoID:         todoID,
			permission:     "admin",
			sharedWithTeam: true,
			expectError:    false,
		},
		{
			name:           "non-member has no permission even when shared",
			userID:         nonMemberID,
			todoID:         todoID,
			permission:     "view",
			sharedWithTeam: true,
			expectError:    true,
			errorCode:      codes.PermissionDenied,
		},
		{
			name:        "non-member has no permission when not shared",
			userID:      nonMemberID,
			todoID:      todoID,
			permission:  "view",
			expectError: true,
			errorCode:   codes.PermissionDenied,
		},
		{
			name:        "empty user ID",
			userID:      "",
			todoID:      todoID,
			permission:  "view",
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name:        "empty TODO ID",
			userID:      ownerID,
			todoID:      "",
			permission:  "view",
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name:        "non-existent TODO",
			userID:      ownerID,
			todoID:      "non-existent-todo",
			permission:  "view",
			expectError: true,
			errorCode:   codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup shared TODO if needed
			if tt.sharedWithTeam {
				todoRepo.sharedTODOs[todoID] = []string{teamID}
			} else {
				delete(todoRepo.sharedTODOs, todoID)
			}

			err := permissionService.CheckTODOPermission(ctx, tt.userID, tt.todoID, tt.permission)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if status, ok := grpcstatus.FromError(err); ok {
					if status.Code() != tt.errorCode {
						t.Errorf("expected error code %v, got %v", tt.errorCode, status.Code())
					}
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestPermissionService_CheckTeamPermission(t *testing.T) {
	ctx := context.Background()
	todoRepo := NewMockTODORepository()
	teamRepo := NewMockTeamRepository()
	permissionService := NewPermissionService(todoRepo, teamRepo)

	// Create test data
	memberID := "member-user"
	adminID := "admin-user"
	ownerID := "owner-user"
	nonMemberID := "non-member-user"
	teamID := "test-team"

	// Create team
	team := &domain.Team{
		ID:        teamID,
		Name:      "Test Team",
		CreatedBy: ownerID,
	}
	teamRepo.teams[teamID] = team

	// Add members to team
	teamRepo.members[teamID] = make(map[string]*domain.TeamMember)
	teamRepo.members[teamID][memberID] = &domain.TeamMember{
		TeamID: teamID,
		UserID: memberID,
		Role:   commonv1.Role_ROLE_MEMBER,
	}
	teamRepo.members[teamID][adminID] = &domain.TeamMember{
		TeamID: teamID,
		UserID: adminID,
		Role:   commonv1.Role_ROLE_ADMIN,
	}
	teamRepo.members[teamID][ownerID] = &domain.TeamMember{
		TeamID: teamID,
		UserID: ownerID,
		Role:   commonv1.Role_ROLE_OWNER,
	}

	tests := []struct {
		name        string
		userID      string
		teamID      string
		permission  string
		expectError bool
		errorCode   codes.Code
	}{
		{
			name:        "member has view permission",
			userID:      memberID,
			teamID:      teamID,
			permission:  "view",
			expectError: false,
		},
		{
			name:        "member has edit permission",
			userID:      memberID,
			teamID:      teamID,
			permission:  "edit",
			expectError: false,
		},
		{
			name:        "admin has admin permission",
			userID:      adminID,
			teamID:      teamID,
			permission:  "admin",
			expectError: false,
		},
		{
			name:        "owner has admin permission",
			userID:      ownerID,
			teamID:      teamID,
			permission:  "admin",
			expectError: false,
		},
		{
			name:        "member has no admin permission",
			userID:      memberID,
			teamID:      teamID,
			permission:  "admin",
			expectError: true,
			errorCode:   codes.PermissionDenied,
		},
		{
			name:        "non-member has no permission",
			userID:      nonMemberID,
			teamID:      teamID,
			permission:  "view",
			expectError: true,
			errorCode:   codes.PermissionDenied,
		},
		{
			name:        "empty user ID",
			userID:      "",
			teamID:      teamID,
			permission:  "view",
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name:        "empty team ID",
			userID:      memberID,
			teamID:      "",
			permission:  "view",
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name:        "non-existent team",
			userID:      memberID,
			teamID:      "non-existent-team",
			permission:  "view",
			expectError: true,
			errorCode:   codes.PermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := permissionService.CheckTeamPermission(ctx, tt.userID, tt.teamID, tt.permission)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if status, ok := grpcstatus.FromError(err); ok {
					if status.Code() != tt.errorCode {
						t.Errorf("expected error code %v, got %v", tt.errorCode, status.Code())
					}
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestPermissionService_hasTeamPermission(t *testing.T) {
	todoRepo := NewMockTODORepository()
	teamRepo := NewMockTeamRepository()
	permissionService := NewPermissionService(todoRepo, teamRepo)

	tests := []struct {
		name               string
		role               commonv1.Role
		permission         string
		expectedPermission bool
	}{
		{
			name:               "member has view permission",
			role:               commonv1.Role_ROLE_MEMBER,
			permission:         "view",
			expectedPermission: true,
		},
		{
			name:               "member has edit permission",
			role:               commonv1.Role_ROLE_MEMBER,
			permission:         "edit",
			expectedPermission: true,
		},
		{
			name:               "member has no admin permission",
			role:               commonv1.Role_ROLE_MEMBER,
			permission:         "admin",
			expectedPermission: false,
		},
		{
			name:               "admin has admin permission",
			role:               commonv1.Role_ROLE_ADMIN,
			permission:         "admin",
			expectedPermission: true,
		},
		{
			name:               "owner has admin permission",
			role:               commonv1.Role_ROLE_OWNER,
			permission:         "admin",
			expectedPermission: true,
		},
		{
			name:               "unknown role has no permission",
			role:               commonv1.Role(999), // Unknown role
			permission:         "view",
			expectedPermission: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasPermission := permissionService.hasTeamPermission(tt.role, tt.permission)

			if hasPermission != tt.expectedPermission {
				t.Errorf("expected permission %v, got %v", tt.expectedPermission, hasPermission)
			}
		})
	}
}
