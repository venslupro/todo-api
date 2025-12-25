package middleware

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/venslupro/todo-api/internal/pkg/auth"
)

const (
	// AuthorizationHeader is the header key for authorization
	AuthorizationHeader = "authorization"
	// BearerPrefix is the prefix for bearer tokens
	BearerPrefix = "Bearer "
)

// Define custom types for context keys to avoid collisions
type contextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
	// UsernameKey is the context key for username
	UsernameKey contextKey = "username"
	// EmailKey is the context key for email
	EmailKey contextKey = "email"
)

// AuthInterceptor creates a gRPC interceptor for authentication
func AuthInterceptor(jwtMgr *auth.JWTManager) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip authentication for certain methods (e.g., health check, login, register)
		if shouldSkipAuth(info.FullMethod) {
			return handler(ctx, req)
		}

		// Extract token from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Get authorization header
		authHeaders := md.Get(AuthorizationHeader)
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		// Extract token
		token := authHeaders[0]
		if !strings.HasPrefix(token, BearerPrefix) {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
		}

		token = strings.TrimPrefix(token, BearerPrefix)

		// Validate token
		claims, err := jwtMgr.Validate(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		// Add user information to context
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UsernameKey, claims.Username)
		ctx = context.WithValue(ctx, EmailKey, claims.Email)

		return handler(ctx, req)
	}
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return "", status.Error(codes.Unauthenticated, "user not authenticated")
	}
	return userID, nil
}

// GetUsernameFromContext extracts username from context
func GetUsernameFromContext(ctx context.Context) string {
	username, ok := ctx.Value(UsernameKey).(string)
	if !ok {
		return ""
	}
	return username
}

// GetEmailFromContext extracts email from context
func GetEmailFromContext(ctx context.Context) string {
	email, ok := ctx.Value(EmailKey).(string)
	if !ok {
		return ""
	}
	return email
}

// shouldSkipAuth determines if authentication should be skipped for a method
func shouldSkipAuth(method string) bool {
	skipMethods := []string{
		"/todo.v1.AuthService/Register",
		"/todo.v1.AuthService/Login",
		"/todo.v1.SystemService/HealthCheck",
	}
	for _, skipMethod := range skipMethods {
		if method == skipMethod {
			return true
		}
	}
	return false
}
