package handlers

import (
	"context"

	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
	todov1 "github.com/venslupro/todo-api/api/gen/todo/v1"
	"github.com/venslupro/todo-api/internal/app/service"
	"github.com/venslupro/todo-api/internal/domain"
	"github.com/venslupro/todo-api/internal/pkg/middleware"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TeamHandler handles team gRPC requests
type TeamHandler struct {
	teamService *service.TeamService
	todov1.UnimplementedTeamServiceServer
}

// NewTeamHandler creates a new team handler
func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

// CreateTeam handles team creation
func (h *TeamHandler) CreateTeam(ctx context.Context, req *todov1.CreateTeamRequest) (*todov1.CreateTeamResponse, error) {
	if req.Name == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "team name is required")
	}

	// Get current user ID from context
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var description string
	if req.Description != nil {
		description = *req.Description
	}

	team, err := h.teamService.CreateTeam(ctx, req.Name, description, userID)
	if err != nil {
		return nil, err
	}

	return &todov1.CreateTeamResponse{
		Team: h.domainTeamToProto(team),
	}, nil
}

// GetTeam handles team retrieval
func (h *TeamHandler) GetTeam(ctx context.Context, req *todov1.GetTeamRequest) (*todov1.GetTeamResponse, error) {
	if req.Id == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "team id is required")
	}

	team, err := h.teamService.GetTeam(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &todov1.GetTeamResponse{
		Team: h.domainTeamToProto(team),
	}, nil
}

// UpdateTeam handles team updates
func (h *TeamHandler) UpdateTeam(ctx context.Context, req *todov1.UpdateTeamRequest) (*todov1.UpdateTeamResponse, error) {
	if req.Id == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "team id is required")
	}

	var name, description string
	if req.Name != nil {
		name = *req.Name
	}
	if req.Description != nil {
		description = *req.Description
	}

	team, err := h.teamService.UpdateTeam(ctx, req.Id, name, description)
	if err != nil {
		return nil, err
	}

	return &todov1.UpdateTeamResponse{
		Team: h.domainTeamToProto(team),
	}, nil
}

// DeleteTeam handles team deletion
func (h *TeamHandler) DeleteTeam(ctx context.Context, req *todov1.DeleteTeamRequest) (*todov1.DeleteTeamResponse, error) {
	if req.Id == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "team id is required")
	}

	err := h.teamService.DeleteTeam(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &todov1.DeleteTeamResponse{}, nil
}

// ListTeams handles listing teams for current user
func (h *TeamHandler) ListTeams(ctx context.Context, req *todov1.ListTeamsRequest) (*todov1.ListTeamsResponse, error) {
	// Get current user ID from context
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	teams, err := h.teamService.ListTeamsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	protoTeams := make([]*todov1.Team, len(teams))
	for i, team := range teams {
		protoTeams[i] = h.domainTeamToProto(team)
	}

	return &todov1.ListTeamsResponse{
		Teams: protoTeams,
	}, nil
}

// AddTeamMember handles adding a member to a team
func (h *TeamHandler) AddTeamMember(ctx context.Context, req *todov1.AddTeamMemberRequest) (*todov1.AddTeamMemberResponse, error) {
	if req.TeamId == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "team id is required")
	}
	if req.UserEmail == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "user email is required")
	}

	var role commonv1.Role
	if req.Role != nil {
		role = *req.Role
	} else {
		role = commonv1.Role_ROLE_MEMBER
	}

	// TODO: Lookup user by email and get user ID
	// For now, we'll use a placeholder
	userID := req.UserEmail // This is temporary

	err := h.teamService.AddTeamMember(ctx, req.TeamId, userID, role)
	if err != nil {
		return nil, err
	}

	// Get the created member
	member, err := h.teamService.GetTeamMember(ctx, req.TeamId, userID)
	if err != nil {
		return nil, err
	}

	return &todov1.AddTeamMemberResponse{
		Member: h.domainTeamMemberToProto(member),
	}, nil
}

// UpdateTeamMember handles updating team member permissions
func (h *TeamHandler) UpdateTeamMember(ctx context.Context, req *todov1.UpdateTeamMemberRequest) (*todov1.UpdateTeamMemberResponse, error) {
	if req.TeamId == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "team id is required")
	}
	if req.UserId == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "user id is required")
	}

	var role commonv1.Role
	if req.Role != nil {
		role = *req.Role
	}

	err := h.teamService.UpdateMemberRole(ctx, req.TeamId, req.UserId, role)
	if err != nil {
		return nil, err
	}

	// Get the updated member
	member, err := h.teamService.GetTeamMember(ctx, req.TeamId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &todov1.UpdateTeamMemberResponse{
		Member: h.domainTeamMemberToProto(member),
	}, nil
}

// RemoveTeamMember handles removing a member from a team
func (h *TeamHandler) RemoveTeamMember(ctx context.Context, req *todov1.RemoveTeamMemberRequest) (*todov1.RemoveTeamMemberResponse, error) {
	if req.TeamId == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "team id is required")
	}
	if req.UserId == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "user id is required")
	}

	err := h.teamService.RemoveTeamMember(ctx, req.TeamId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &todov1.RemoveTeamMemberResponse{}, nil
}

// ListTeamMembers handles listing team members
func (h *TeamHandler) ListTeamMembers(ctx context.Context, req *todov1.ListTeamMembersRequest) (*todov1.ListTeamMembersResponse, error) {
	if req.TeamId == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "team id is required")
	}

	members, err := h.teamService.ListTeamMembers(ctx, req.TeamId)
	if err != nil {
		return nil, err
	}

	protoMembers := make([]*todov1.TeamMember, len(members))
	for i, member := range members {
		protoMembers[i] = h.domainTeamMemberToProto(member)
	}

	return &todov1.ListTeamMembersResponse{
		Members: protoMembers,
	}, nil
}

// ShareList handles sharing a TODO list with a team
func (h *TeamHandler) ShareList(ctx context.Context, req *todov1.ShareListRequest) (*todov1.ShareListResponse, error) {
	if req.TodoId == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "todo id is required")
	}
	if req.TeamId == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "team id is required")
	}

	// Get current user ID from context
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = h.teamService.ShareTODO(ctx, req.TodoId, req.TeamId, userID)
	if err != nil {
		return nil, err
	}

	// TODO: Create and return SharedList object
	return &todov1.ShareListResponse{
		SharedList: &todov1.SharedList{
			TodoId:   req.TodoId,
			TeamId:   req.TeamId,
			SharedBy: userID,
			SharedAt: timestamppb.Now(),
		},
	}, nil
}

// domainTeamToProto converts domain Team to protobuf Team
func (h *TeamHandler) domainTeamToProto(team *domain.Team) *todov1.Team {
	return &todov1.Team{
		Id:          team.ID,
		Name:        team.Name,
		Description: team.Description,
		OwnerId:     team.CreatedBy,
		CreatedAt:   timestamppb.New(team.CreatedAt),
		UpdatedAt:   timestamppb.New(team.UpdatedAt),
	}
}

// domainTeamMemberToProto converts domain TeamMember to protobuf TeamMember
func (h *TeamHandler) domainTeamMemberToProto(member *domain.TeamMember) *todov1.TeamMember {
	return &todov1.TeamMember{
		TeamId:   member.TeamID,
		UserId:   member.UserID,
		Role:     member.Role,
		JoinedAt: timestamppb.New(member.JoinedAt),
	}
}
