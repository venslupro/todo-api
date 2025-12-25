package middleware

import (
	"context"
	"testing"

	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
	todov1 "github.com/venslupro/todo-api/api/gen/todo/v1"
	"github.com/venslupro/todo-api/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockTeamRepository is a mock implementation of TeamRepository for testing
type MockTeamRepository struct {
	members map[string]*domain.TeamMember
}

func NewMockTeamRepository() *MockTeamRepository {
	return &MockTeamRepository{
		members: make(map[string]*domain.TeamMember),
	}
}

func (m *MockTeamRepository) Create(ctx context.Context, team *domain.Team) error {
	return nil
}

func (m *MockTeamRepository) GetByID(ctx context.Context, id string) (*domain.Team, error) {
	return nil, nil
}

func (m *MockTeamRepository) Update(ctx context.Context, team *domain.Team) error {
	return nil
}

func (m *MockTeamRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *MockTeamRepository) ListByUser(ctx context.Context, userID string) ([]*domain.Team, error) {
	return nil, nil
}

func (m *MockTeamRepository) AddMember(ctx context.Context, member *domain.TeamMember) error {
	key := member.TeamID + "_" + member.UserID
	m.members[key] = member
	return nil
}

func (m *MockTeamRepository) RemoveMember(ctx context.Context, teamID, userID string) error {
	key := teamID + "_" + userID
	delete(m.members, key)
	return nil
}

func (m *MockTeamRepository) GetMember(ctx context.Context, teamID, userID string) (*domain.TeamMember, error) {
	key := teamID + "_" + userID
	member, ok := m.members[key]
	if !ok {
		return nil, status.Error(codes.NotFound, "member not found")
	}
	return member, nil
}

func (m *MockTeamRepository) ListMembers(ctx context.Context, teamID string) ([]*domain.TeamMember, error) {
	var members []*domain.TeamMember
	for key, member := range m.members {
		if len(key) > len(teamID) && key[:len(teamID)] == teamID {
			members = append(members, member)
		}
	}
	return members, nil
}

func (m *MockTeamRepository) UpdateMemberRole(ctx context.Context, teamID, userID string, role commonv1.Role) error {
	key := teamID + "_" + userID
	member, ok := m.members[key]
	if !ok {
		return status.Error(codes.NotFound, "member not found")
	}
	member.Role = role
	return nil
}

func (m *MockTeamRepository) ShareTODO(ctx context.Context, todoID, teamID, sharedBy string) error {
	return nil
}

func (m *MockTeamRepository) UnshareTODO(ctx context.Context, todoID, teamID string) error {
	return nil
}

func (m *MockTeamRepository) GetSharedTODOs(ctx context.Context, teamID string) ([]string, error) {
	return nil, nil
}

func (m *MockTeamRepository) GetSharedTeams(ctx context.Context, todoID string) ([]string, error) {
	return nil, nil
}

func TestAuthorizationInterceptor(t *testing.T) {
	teamRepo := NewMockTeamRepository()
	interceptor := AuthorizationInterceptor(teamRepo)

	tests := []struct {
		name       string
		fullMethod string
		setupCtx   func() context.Context
		setupReq   func() interface{}
		wantErr    bool
		errorCode  codes.Code
	}{
		{
			name:       "skip auth for register method",
			fullMethod: "/todo.v1.AuthService/Register",
			setupCtx:   func() context.Context { return context.Background() },
			setupReq:   func() interface{} { return nil },
			wantErr:    false,
		},
		{
			name:       "skip auth for login method",
			fullMethod: "/todo.v1.AuthService/Login",
			setupCtx:   func() context.Context { return context.Background() },
			setupReq:   func() interface{} { return nil },
			wantErr:    false,
		},
		{
			name:       "method without required permission",
			fullMethod: "/todo.v1.TODOService/GetProfile",
			setupCtx:   func() context.Context { return context.WithValue(context.Background(), UserIDKey, "user-123") },
			setupReq:   func() interface{} { return nil },
			wantErr:    false,
		},
		{
			name:       "TODO operation without resource ID",
			fullMethod: "/todo.v1.TODOService/CreateTODO",
			setupCtx:   func() context.Context { return context.WithValue(context.Background(), UserIDKey, "user-123") },
			setupReq:   func() interface{} { return &todov1.CreateTODORequest{Title: "Test TODO"} },
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			req := tt.setupReq()
			info := &grpc.UnaryServerInfo{
				FullMethod: tt.fullMethod,
			}

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return "success", nil
			}

			_, err := interceptor(ctx, req, info, handler)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
					return
				}

				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("error is not a gRPC status error: %v", err)
					return
				}

				if st.Code() != tt.errorCode {
					t.Errorf("error code = %v, want %v", st.Code(), tt.errorCode)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestShouldSkipAuthorization(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		wantResult bool
	}{
		{
			name:       "register method",
			method:     "/todo.v1.AuthService/Register",
			wantResult: true,
		},
		{
			name:       "login method",
			method:     "/todo.v1.AuthService/Login",
			wantResult: true,
		},
		{
			name:       "logout method",
			method:     "/todo.v1.AuthService/Logout",
			wantResult: true,
		},
		{
			name:       "health check method",
			method:     "/todo.v1.SystemService/HealthCheck",
			wantResult: true,
		},
		{
			name:       "get profile method",
			method:     "/todo.v1.TODOService/GetProfile",
			wantResult: true,
		},
		{
			name:       "other method",
			method:     "/todo.v1.TODOService/CreateTODO",
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSkipAuthorization(tt.method)
			if result != tt.wantResult {
				t.Errorf("shouldSkipAuthorization(%v) = %v, want %v", tt.method, result, tt.wantResult)
			}
		})
	}
}

func TestGetRequiredPermission(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		wantResult string
	}{
		{
			name:       "create TODO method",
			method:     "/todo.v1.TODOService/CreateTODO",
			wantResult: PermissionEdit,
		},
		{
			name:       "get TODO method",
			method:     "/todo.v1.TODOService/GetTODO",
			wantResult: PermissionView,
		},
		{
			name:       "create team method",
			method:     "/todo.v1.TeamService/CreateTeam",
			wantResult: PermissionAdmin,
		},
		{
			name:       "unknown method",
			method:     "/todo.v1.UnknownService/Method",
			wantResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getRequiredPermission(tt.method)
			if result != tt.wantResult {
				t.Errorf("getRequiredPermission(%v) = %v, want %v", tt.method, result, tt.wantResult)
			}
		})
	}
}

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name               string
		role               commonv1.Role
		requiredPermission string
		wantResult         bool
	}{
		{
			name:               "member has view permission",
			role:               commonv1.Role_ROLE_MEMBER,
			requiredPermission: PermissionView,
			wantResult:         true,
		},
		{
			name:               "member has edit permission",
			role:               commonv1.Role_ROLE_MEMBER,
			requiredPermission: PermissionEdit,
			wantResult:         true,
		},
		{
			name:               "member does not have admin permission",
			role:               commonv1.Role_ROLE_MEMBER,
			requiredPermission: PermissionAdmin,
			wantResult:         false,
		},
		{
			name:               "admin has all permissions",
			role:               commonv1.Role_ROLE_ADMIN,
			requiredPermission: PermissionAdmin,
			wantResult:         true,
		},
		{
			name:               "owner has all permissions",
			role:               commonv1.Role_ROLE_OWNER,
			requiredPermission: PermissionAdmin,
			wantResult:         true,
		},
		{
			name:               "unknown role has no permissions",
			role:               commonv1.Role_ROLE_UNSPECIFIED,
			requiredPermission: PermissionView,
			wantResult:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasPermission(tt.role, tt.requiredPermission)
			if result != tt.wantResult {
				t.Errorf("hasPermission(%v, %v) = %v, want %v", tt.role, tt.requiredPermission, result, tt.wantResult)
			}
		})
	}
}

func TestIsAdmin(t *testing.T) {
	teamRepo := NewMockTeamRepository()

	// Setup test member
	member := &domain.TeamMember{
		TeamID: "team-123",
		UserID: "user-admin",
		Role:   commonv1.Role_ROLE_ADMIN,
	}
	teamRepo.AddMember(context.Background(), member)

	tests := []struct {
		name       string
		setupCtx   func() context.Context
		teamID     string
		wantResult bool
	}{
		{
			name: "admin user is admin",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), UserIDKey, "user-admin")
			},
			teamID:     "team-123",
			wantResult: true,
		},
		{
			name: "unknown user is not admin",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), UserIDKey, "user-unknown")
			},
			teamID:     "team-123",
			wantResult: false,
		},
		{
			name: "missing user ID",
			setupCtx: func() context.Context {
				return context.Background()
			},
			teamID:     "team-123",
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			result := IsAdmin(ctx, tt.teamID, teamRepo)
			if result != tt.wantResult {
				t.Errorf("IsAdmin() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}

func TestIsOwner(t *testing.T) {
	teamRepo := NewMockTeamRepository()

	// Setup test member
	member := &domain.TeamMember{
		TeamID: "team-123",
		UserID: "user-owner",
		Role:   commonv1.Role_ROLE_OWNER,
	}
	teamRepo.AddMember(context.Background(), member)

	tests := []struct {
		name       string
		setupCtx   func() context.Context
		teamID     string
		wantResult bool
	}{
		{
			name: "owner user is owner",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), UserIDKey, "user-owner")
			},
			teamID:     "team-123",
			wantResult: true,
		},
		{
			name: "unknown user is not owner",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), UserIDKey, "user-unknown")
			},
			teamID:     "team-123",
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			result := IsOwner(ctx, tt.teamID, teamRepo)
			if result != tt.wantResult {
				t.Errorf("IsOwner() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}
