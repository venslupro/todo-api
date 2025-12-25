package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/venslupro/todo-api/internal/app/service"
	"github.com/venslupro/todo-api/internal/domain"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

// WebSocketHandler handles WebSocket connections for real-time updates
type WebSocketHandler struct {
	websocketService *service.WebSocketService
	authService      *service.AuthService
	teamService      *service.TeamService
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(
	websocketService *service.WebSocketService,
	authService *service.AuthService,
	teamService *service.TeamService,
) *WebSocketHandler {
	return &WebSocketHandler{
		websocketService: websocketService,
		authService:      authService,
		teamService:      teamService,
	}
}

// HandleWebSocket handles WebSocket connection requests
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract JWT token from query parameters or headers
	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.Header.Get("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}

	if token == "" {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Validate JWT token
	claims, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	userID := claims.UserID
	if userID == "" {
		http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
		return
	}

	// Get user's team memberships
	teamIDs, err := h.getUserTeamIDs(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get user teams", http.StatusInternalServerError)
		return
	}

	// Handle WebSocket connection
	err = h.websocketService.HandleWebSocketConnection(w, r, userID, teamIDs)
	if err != nil {
		http.Error(w, fmt.Sprintf("WebSocket connection failed: %v", err), http.StatusInternalServerError)
		return
	}
}

// getUserTeamIDs retrieves the team IDs that a user belongs to
func (h *WebSocketHandler) getUserTeamIDs(ctx context.Context, userID string) ([]string, error) {
	teams, err := h.teamService.ListTeamsByUser(ctx, userID)
	if err != nil {
		// If user has no teams, return empty slice
		if grpcstatus.Code(err) == codes.NotFound {
			return []string{}, nil
		}
		return nil, err
	}

	teamIDs := make([]string, len(teams))
	for i, team := range teams {
		teamIDs[i] = team.ID
	}

	return teamIDs, nil
}

// BroadcastTODOUpdate broadcasts a TODO update to relevant clients
func (h *WebSocketHandler) BroadcastTODOUpdate(ctx context.Context, todo *domain.TODO, action string) error {
	h.websocketService.BroadcastTODOUpdate(ctx, todo, action)
	return nil
}

// BroadcastTeamUpdate broadcasts a team update to team members
func (h *WebSocketHandler) BroadcastTeamUpdate(ctx context.Context, team *domain.Team, action string) error {
	h.websocketService.BroadcastTeamUpdate(ctx, team, action)
	return nil
}

// BroadcastUserNotification broadcasts a notification to a specific user
func (h *WebSocketHandler) BroadcastUserNotification(ctx context.Context, userID, messageType, content string) error {
	h.websocketService.BroadcastUserNotification(ctx, userID, messageType, content)
	return nil
}
