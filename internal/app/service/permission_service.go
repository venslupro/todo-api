package service

import (
	"context"
	"fmt"

	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
	"github.com/venslupro/todo-api/internal/domain"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

// PermissionService provides authorization and permission checking services
type PermissionService struct {
	todoRepo domain.TODORepository
	teamRepo domain.TeamRepository
}

// NewPermissionService creates a new permission service
func NewPermissionService(todoRepo domain.TODORepository, teamRepo domain.TeamRepository) *PermissionService {
	return &PermissionService{
		todoRepo: todoRepo,
		teamRepo: teamRepo,
	}
}

// CheckTODOPermission checks if a user has permission to access a TODO
func (s *PermissionService) CheckTODOPermission(ctx context.Context, userID, todoID, requiredPermission string) error {
	if userID == "" || todoID == "" {
		return grpcstatus.Error(codes.InvalidArgument, "user id and todo id are required")
	}

	// Get the TODO
	todo, err := s.todoRepo.GetByID(ctx, todoID)
	if err != nil {
		return grpcstatus.Error(codes.NotFound, fmt.Sprintf("todo not found: %v", err))
	}

	// Check ownership
	if todo.UserID == userID {
		// Owner has full permissions
		return nil
	}

	// Check if TODO is shared with any teams the user belongs to
	sharedTeams, err := s.todoRepo.GetSharedTeams(ctx, todoID)
	if err != nil {
		return grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to check shared teams: %v", err))
	}

	for _, teamID := range sharedTeams {
		// Check if user is a member of this team
		member, err := s.teamRepo.GetMember(ctx, teamID, userID)
		if err != nil {
			continue // User is not a member of this team
		}

		// Check if member has required permission
		if s.hasTeamPermission(member.Role, requiredPermission) {
			return nil
		}
	}

	return grpcstatus.Error(codes.PermissionDenied, "insufficient permissions to access this todo")
}

// CheckTeamPermission checks if a user has permission to access a team
func (s *PermissionService) CheckTeamPermission(ctx context.Context, userID, teamID, requiredPermission string) error {
	if userID == "" || teamID == "" {
		return grpcstatus.Error(codes.InvalidArgument, "user id and team id are required")
	}

	// Get team member
	member, err := s.teamRepo.GetMember(ctx, teamID, userID)
	if err != nil {
		return grpcstatus.Error(codes.PermissionDenied, "user is not a member of this team")
	}

	// Check if member has required permission
	if !s.hasTeamPermission(member.Role, requiredPermission) {
		return grpcstatus.Error(codes.PermissionDenied, "insufficient permissions")
	}

	return nil
}

// CanCreateTODOInTeam checks if a user can create a TODO in a team
func (s *PermissionService) CanCreateTODOInTeam(ctx context.Context, userID, teamID string) error {
	return s.CheckTeamPermission(ctx, userID, teamID, "edit")
}

// CanViewTODO checks if a user can view a TODO
func (s *PermissionService) CanViewTODO(ctx context.Context, userID, todoID string) error {
	return s.CheckTODOPermission(ctx, userID, todoID, "view")
}

// CanEditTODO checks if a user can edit a TODO
func (s *PermissionService) CanEditTODO(ctx context.Context, userID, todoID string) error {
	return s.CheckTODOPermission(ctx, userID, todoID, "edit")
}

// CanDeleteTODO checks if a user can delete a TODO
func (s *PermissionService) CanDeleteTODO(ctx context.Context, userID, todoID string) error {
	return s.CheckTODOPermission(ctx, userID, todoID, "edit")
}

// CanManageTeam checks if a user can manage a team (admin permissions)
func (s *PermissionService) CanManageTeam(ctx context.Context, userID, teamID string) error {
	return s.CheckTeamPermission(ctx, userID, teamID, "admin")
}

// IsTeamOwner checks if a user is the owner of a team
func (s *PermissionService) IsTeamOwner(ctx context.Context, userID, teamID string) (bool, error) {
	member, err := s.teamRepo.GetMember(ctx, teamID, userID)
	if err != nil {
		return false, err
	}

	return member.Role == commonv1.Role_ROLE_OWNER, nil
}

// IsTeamAdmin checks if a user is an admin or owner of a team
func (s *PermissionService) IsTeamAdmin(ctx context.Context, userID, teamID string) (bool, error) {
	member, err := s.teamRepo.GetMember(ctx, teamID, userID)
	if err != nil {
		return false, err
	}

	return member.Role == commonv1.Role_ROLE_ADMIN || member.Role == commonv1.Role_ROLE_OWNER, nil
}

// hasTeamPermission checks if a role has the required permission
func (s *PermissionService) hasTeamPermission(role commonv1.Role, requiredPermission string) bool {
	rolePermissions := map[commonv1.Role][]string{
		commonv1.Role_ROLE_MEMBER: {"view", "edit"},
		commonv1.Role_ROLE_ADMIN:  {"view", "edit", "admin"},
		commonv1.Role_ROLE_OWNER:  {"view", "edit", "admin"},
	}

	permissions, exists := rolePermissions[role]
	if !exists {
		return false
	}

	for _, permission := range permissions {
		if permission == requiredPermission {
			return true
		}
	}

	return false
}

// GetUserTeams retrieves all teams a user belongs to
func (s *PermissionService) GetUserTeams(ctx context.Context, userID string) ([]*domain.Team, error) {
	return s.teamRepo.ListByUser(ctx, userID)
}

// GetTeamMembers retrieves all members of a team
func (s *PermissionService) GetTeamMembers(ctx context.Context, teamID string) ([]*domain.TeamMember, error) {
	return s.teamRepo.ListMembers(ctx, teamID)
}

// UpdateMemberRole updates a team member's role
func (s *PermissionService) UpdateMemberRole(ctx context.Context, teamID, userID string, role commonv1.Role) error {
	return s.teamRepo.UpdateMemberRole(ctx, teamID, userID, role)
}

// ShareTODOWithTeam shares a TODO with a team
func (s *PermissionService) ShareTODOWithTeam(ctx context.Context, todoID, teamID, sharedBy string) error {
	return s.teamRepo.ShareTODO(ctx, todoID, teamID, sharedBy)
}

// UnshareTODOFromTeam unshares a TODO from a team
func (s *PermissionService) UnshareTODOFromTeam(ctx context.Context, todoID, teamID string) error {
	return s.teamRepo.UnshareTODO(ctx, todoID, teamID)
}
