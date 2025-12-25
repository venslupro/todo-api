package middleware

import (
	"context"
	"reflect"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
	todov1 "github.com/venslupro/todo-api/api/gen/todo/v1"
	"github.com/venslupro/todo-api/internal/domain"
)

// Permission represents the required permission for a method
const (
	PermissionView  = "view"
	PermissionEdit  = "edit"
	PermissionAdmin = "admin"
)

// MethodPermission defines the required permission for each gRPC method
type MethodPermission struct {
	Method     string
	Permission string
}

// AuthorizationInterceptor creates a gRPC interceptor for authorization
func AuthorizationInterceptor(teamRepo domain.TeamRepository) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip authorization for public methods
		if shouldSkipAuthorization(info.FullMethod) {
			return handler(ctx, req)
		}

		// Get user ID from context
		userID, err := GetUserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}

		// Get required permission for the method
		requiredPermission := getRequiredPermission(info.FullMethod)
		if requiredPermission == "" {
			// No specific permission required, allow access
			return handler(ctx, req)
		}

		// Extract resource ID from request if needed
		resourceID, teamID := extractResourceInfo(req, info.FullMethod)

		// Check authorization
		if err := checkAuthorization(ctx, userID, resourceID, teamID, requiredPermission, teamRepo); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// shouldSkipAuthorization determines if authorization should be skipped for a method
func shouldSkipAuthorization(method string) bool {
	skipMethods := []string{
		"/todo.v1.AuthService/Register",
		"/todo.v1.AuthService/Login",
		"/todo.v1.AuthService/Logout",
		"/todo.v1.AuthService/RefreshToken",
		"/todo.v1.SystemService/HealthCheck",
		"/todo.v1.TODOService/GetProfile",
	}
	for _, skipMethod := range skipMethods {
		if method == skipMethod {
			return true
		}
	}
	return false
}

// getRequiredPermission returns the required permission for a method
func getRequiredPermission(method string) string {
	methodPermissions := map[string]string{
		// TODO operations
		"/todo.v1.TODOService/CreateTODO": PermissionEdit,
		"/todo.v1.TODOService/GetTODO":    PermissionView,
		"/todo.v1.TODOService/UpdateTODO": PermissionEdit,
		"/todo.v1.TODOService/DeleteTODO": PermissionEdit,
		"/todo.v1.TODOService/ListTODOs":  PermissionView,

		// Team operations
		"/todo.v1.TeamService/CreateTeam":       PermissionAdmin,
		"/todo.v1.TeamService/GetTeam":          PermissionView,
		"/todo.v1.TeamService/UpdateTeam":       PermissionAdmin,
		"/todo.v1.TeamService/DeleteTeam":       PermissionAdmin,
		"/todo.v1.TeamService/ListTeams":        PermissionView,
		"/todo.v1.TeamService/AddTeamMember":    PermissionAdmin,
		"/todo.v1.TeamService/RemoveTeamMember": PermissionAdmin,

		// Media operations
		"/todo.v1.MediaService/UploadMedia": PermissionEdit,
		"/todo.v1.MediaService/GetMedia":    PermissionView,
		"/todo.v1.MediaService/DeleteMedia": PermissionEdit,

		// Real-time operations
		"/todo.v1.RealtimeService/Subscribe":    PermissionView,
		"/todo.v1.RealtimeService/PublishEvent": PermissionEdit,
	}

	return methodPermissions[method]
}

// extractResourceInfo extracts resource ID and team ID from request
func extractResourceInfo(req interface{}, method string) (resourceID, teamID string) {
	// Use reflection to extract resource IDs from request objects
	v := reflect.ValueOf(req)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return "", ""
	}

	// Extract TODO ID for TODO operations
	if strings.Contains(method, "TODOService") {
		if field := v.FieldByName("Id"); field.IsValid() && field.Kind() == reflect.String {
			resourceID = field.String()
		}
		if field := v.FieldByName("TodoId"); field.IsValid() && field.Kind() == reflect.String {
			resourceID = field.String()
		}
	}

	// Extract team ID for team operations
	if strings.Contains(method, "TeamService") {
		if field := v.FieldByName("Id"); field.IsValid() && field.Kind() == reflect.String {
			teamID = field.String()
		}
		if field := v.FieldByName("TeamId"); field.IsValid() && field.Kind() == reflect.String {
			teamID = field.String()
		}
	}

	// Extract resource ID for specific methods
	switch req := req.(type) {
	case *todov1.GetTODORequest:
		resourceID = req.Id
	case *todov1.UpdateTODORequest:
		resourceID = req.Id
	case *todov1.DeleteTODORequest:
		resourceID = req.Id
	}

	return resourceID, teamID
}

// checkAuthorization checks if user has required permission
func checkAuthorization(ctx context.Context, userID, resourceID, teamID, requiredPermission string, teamRepo domain.TeamRepository) error {
	// If no specific resource, check global permissions
	if resourceID == "" && teamID == "" {
		// For now, allow all authenticated users to perform global operations
		return nil
	}

	// Check team permissions if team ID is provided
	if teamID != "" {
		member, err := teamRepo.GetMember(ctx, teamID, userID)
		if err != nil {
			return status.Error(codes.PermissionDenied, "user is not a member of this team")
		}

		if !hasPermission(member.Role, requiredPermission) {
			return status.Error(codes.PermissionDenied, "insufficient permissions")
		}
		return nil
	}

	// For TODO operations, check ownership or team membership
	if resourceID != "" {
		// In a real implementation, you would check if the user owns the TODO
		// or has access through team membership
		// For now, we'll assume ownership check is handled at service level
		return nil
	}

	return status.Error(codes.PermissionDenied, "access denied")
}

// hasPermission checks if a role has the required permission
func hasPermission(role commonv1.Role, requiredPermission string) bool {
	rolePermissions := map[commonv1.Role][]string{
		commonv1.Role_ROLE_MEMBER: {PermissionView, PermissionEdit},
		commonv1.Role_ROLE_ADMIN:  {PermissionView, PermissionEdit, PermissionAdmin},
		commonv1.Role_ROLE_OWNER:  {PermissionView, PermissionEdit, PermissionAdmin},
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

// IsAdmin checks if the current user has admin role in a team
func IsAdmin(ctx context.Context, teamID string, teamRepo domain.TeamRepository) bool {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return false
	}

	member, err := teamRepo.GetMember(ctx, teamID, userID)
	if err != nil {
		return false
	}

	return member.Role == commonv1.Role_ROLE_ADMIN || member.Role == commonv1.Role_ROLE_OWNER
}

// IsOwner checks if the current user is the owner of a team
func IsOwner(ctx context.Context, teamID string, teamRepo domain.TeamRepository) bool {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return false
	}

	member, err := teamRepo.GetMember(ctx, teamID, userID)
	if err != nil {
		return false
	}

	return member.Role == commonv1.Role_ROLE_OWNER
}
