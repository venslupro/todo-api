package service

import (
	"context"
	"fmt"

	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
	"github.com/venslupro/todo-api/internal/domain"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

// TeamService provides business logic for team operations
type TeamService struct {
	teamRepo         domain.TeamRepository
	websocketService *WebSocketService
}

// NewTeamService creates a new team service
func NewTeamService(teamRepo domain.TeamRepository, websocketService *WebSocketService) *TeamService {
	return &TeamService{
		teamRepo:         teamRepo,
		websocketService: websocketService,
	}
}

// CreateTeam creates a new team
func (s *TeamService) CreateTeam(ctx context.Context, name, description, createdBy string) (*domain.Team, error) {
	if name == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "team name is required")
	}
	if createdBy == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "created by user id is required")
	}

	team := domain.NewTeam(name, description, createdBy)

	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to create team: %v", err))
	}

	// Add creator as team owner
	member := &domain.TeamMember{
		TeamID:   team.ID,
		UserID:   createdBy,
		Role:     commonv1.Role_ROLE_OWNER,
		JoinedAt: team.CreatedAt,
	}

	if err := s.teamRepo.AddMember(ctx, member); err != nil {
		// If adding member fails, delete the team
		_ = s.teamRepo.Delete(ctx, team.ID)
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to add team owner: %v", err))
	}

	// Broadcast WebSocket notification
	if s.websocketService != nil {
		s.websocketService.BroadcastTeamUpdate(ctx, team, "created")
	}

	return team, nil
}

// GetTeam retrieves a team by ID
func (s *TeamService) GetTeam(ctx context.Context, teamID string) (*domain.Team, error) {
	if teamID == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "team id is required")
	}

	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return nil, grpcstatus.Error(codes.NotFound, fmt.Sprintf("team not found: %v", err))
	}

	return team, nil
}

// UpdateTeam updates an existing team
func (s *TeamService) UpdateTeam(ctx context.Context, teamID, name, description string) (*domain.Team, error) {
	if teamID == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "team id is required")
	}

	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return nil, grpcstatus.Error(codes.NotFound, fmt.Sprintf("team not found: %v", err))
	}

	if name != "" {
		team.Name = name
	}
	if description != "" {
		team.Description = description
	}

	if err := s.teamRepo.Update(ctx, team); err != nil {
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to update team: %v", err))
	}

	// Broadcast WebSocket notification
	if s.websocketService != nil {
		s.websocketService.BroadcastTeamUpdate(ctx, team, "updated")
	}

	return team, nil
}

// DeleteTeam deletes a team
func (s *TeamService) DeleteTeam(ctx context.Context, teamID string) error {
	if teamID == "" {
		return grpcstatus.Error(codes.InvalidArgument, "team id is required")
	}

	// Get team before deletion for WebSocket notification
	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil || team == nil {
		return grpcstatus.Error(codes.NotFound, "team not found")
	}

	if err := s.teamRepo.Delete(ctx, teamID); err != nil {
		return grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to delete team: %v", err))
	}

	// Broadcast WebSocket notification
	if s.websocketService != nil {
		s.websocketService.BroadcastTeamUpdate(ctx, team, "deleted")
	}

	return nil
}

// GetTeamMember retrieves a team member by team ID and user ID
func (s *TeamService) GetTeamMember(ctx context.Context, teamID, userID string) (*domain.TeamMember, error) {
	if teamID == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "team id is required")
	}
	if userID == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "user id is required")
	}

	member, err := s.teamRepo.GetMember(ctx, teamID, userID)
	if err != nil {
		return nil, grpcstatus.Error(codes.NotFound, fmt.Sprintf("team member not found: %v", err))
	}

	return member, nil
}

// ListTeamsByUser retrieves teams for a user
func (s *TeamService) ListTeamsByUser(ctx context.Context, userID string) ([]*domain.Team, error) {
	if userID == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "user id is required")
	}

	teams, err := s.teamRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to list teams: %v", err))
	}

	return teams, nil
}

// AddTeamMember adds a member to a team
func (s *TeamService) AddTeamMember(ctx context.Context, teamID, userID string, role commonv1.Role) error {
	if teamID == "" {
		return grpcstatus.Error(codes.InvalidArgument, "team id is required")
	}
	if userID == "" {
		return grpcstatus.Error(codes.InvalidArgument, "user id is required")
	}

	// Check if team exists
	team, err := s.teamRepo.GetByID(ctx, teamID)
	if err != nil {
		return grpcstatus.Error(codes.NotFound, "team not found")
	}

	// Check if user is already a member
	existingMember, err := s.teamRepo.GetMember(ctx, teamID, userID)
	if err == nil && existingMember != nil {
		return grpcstatus.Error(codes.AlreadyExists, "user is already a team member")
	}

	member := &domain.TeamMember{
		TeamID:   teamID,
		UserID:   userID,
		Role:     role,
		JoinedAt: team.CreatedAt,
	}

	if err := s.teamRepo.AddMember(ctx, member); err != nil {
		return grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to add team member: %v", err))
	}

	return nil
}

// RemoveTeamMember removes a member from a team
func (s *TeamService) RemoveTeamMember(ctx context.Context, teamID, userID string) error {
	if teamID == "" {
		return grpcstatus.Error(codes.InvalidArgument, "team id is required")
	}
	if userID == "" {
		return grpcstatus.Error(codes.InvalidArgument, "user id is required")
	}

	// Check if member exists
	member, err := s.teamRepo.GetMember(ctx, teamID, userID)
	if err != nil || member == nil {
		return grpcstatus.Error(codes.NotFound, "team member not found")
	}

	if err := s.teamRepo.RemoveMember(ctx, teamID, userID); err != nil {
		return grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to remove team member: %v", err))
	}

	return nil
}

// UpdateMemberRole updates a member's role
func (s *TeamService) UpdateMemberRole(ctx context.Context, teamID, userID string, role commonv1.Role) error {
	if teamID == "" {
		return grpcstatus.Error(codes.InvalidArgument, "team id is required")
	}
	if userID == "" {
		return grpcstatus.Error(codes.InvalidArgument, "user id is required")
	}

	// Check if member exists
	member, err := s.teamRepo.GetMember(ctx, teamID, userID)
	if err != nil || member == nil {
		return grpcstatus.Error(codes.NotFound, "team member not found")
	}

	if err := s.teamRepo.UpdateMemberRole(ctx, teamID, userID, role); err != nil {
		return grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to update member role: %v", err))
	}

	return nil
}

// ListTeamMembers retrieves all members of a team
func (s *TeamService) ListTeamMembers(ctx context.Context, teamID string) ([]*domain.TeamMember, error) {
	if teamID == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "team id is required")
	}

	members, err := s.teamRepo.ListMembers(ctx, teamID)
	if err != nil {
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to list team members: %v", err))
	}

	return members, nil
}

// ShareTODO shares a TODO with a team
func (s *TeamService) ShareTODO(ctx context.Context, todoID, teamID, sharedBy string) error {
	if todoID == "" {
		return grpcstatus.Error(codes.InvalidArgument, "todo id is required")
	}
	if teamID == "" {
		return grpcstatus.Error(codes.InvalidArgument, "team id is required")
	}
	if sharedBy == "" {
		return grpcstatus.Error(codes.InvalidArgument, "shared by user id is required")
	}

	if err := s.teamRepo.ShareTODO(ctx, todoID, teamID, sharedBy); err != nil {
		return grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to share todo: %v", err))
	}

	return nil
}

// UnshareTODO unshares a TODO from a team
func (s *TeamService) UnshareTODO(ctx context.Context, todoID, teamID string) error {
	if todoID == "" {
		return grpcstatus.Error(codes.InvalidArgument, "todo id is required")
	}
	if teamID == "" {
		return grpcstatus.Error(codes.InvalidArgument, "team id is required")
	}

	if err := s.teamRepo.UnshareTODO(ctx, todoID, teamID); err != nil {
		return grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to unshare todo: %v", err))
	}

	return nil
}
