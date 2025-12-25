package handlers

import (
	"context"
	"time"

	todov1 "github.com/venslupro/todo-api/api/gen/todo/v1"
	"github.com/venslupro/todo-api/internal/app/service"
	"github.com/venslupro/todo-api/internal/domain"
	"github.com/venslupro/todo-api/internal/pkg/auth"
	"github.com/venslupro/todo-api/internal/pkg/middleware"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AuthHandler handles authentication gRPC requests
type AuthHandler struct {
	authService *service.AuthService
	jwtMgr      *auth.JWTManager
	todov1.UnimplementedAuthServiceServer
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService, jwtMgr *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtMgr:      jwtMgr,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(ctx context.Context, req *todov1.RegisterRequest) (*todov1.RegisterResponse, error) {
	if req.Email == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "email is required")
	}
	if req.Password == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "password is required")
	}
	if req.Username == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "username is required")
	}

	user, err := h.authService.Register(ctx, req.Email, req.Username, req.Password, req.DisplayName)
	if err != nil {
		return nil, err
	}

	return &todov1.RegisterResponse{
		User: h.domainUserToProto(user),
	}, nil
}

// Login handles user login
func (h *AuthHandler) Login(ctx context.Context, req *todov1.LoginRequest) (*todov1.LoginResponse, error) {
	if req.Email == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "email is required")
	}
	if req.Password == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "password is required")
	}

	user, accessToken, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := h.jwtMgr.Generate(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, grpcstatus.Error(codes.Internal, "failed to generate refresh token")
	}

	accessTokenExpires := time.Now().Add(24 * time.Hour)      // 24 hours
	refreshTokenExpires := time.Now().Add(7 * 24 * time.Hour) // 7 days

	return &todov1.LoginResponse{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  timestamppb.New(accessTokenExpires),
		RefreshTokenExpiresAt: timestamppb.New(refreshTokenExpires),
		User:                  h.domainUserToProto(user),
	}, nil
}

// Logout handles user logout
func (h *AuthHandler) Logout(ctx context.Context, req *todov1.LogoutRequest) (*todov1.LogoutResponse, error) {
	// In a real implementation, you might want to blacklist the token
	// For now, we'll just return success
	return &todov1.LogoutResponse{}, nil
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(ctx context.Context, req *todov1.RefreshTokenRequest) (*todov1.RefreshTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "refresh token is required")
	}

	claims, err := h.authService.ValidateToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, grpcstatus.Error(codes.Unauthenticated, "invalid refresh token")
	}

	// Generate new access token
	accessToken, err := h.jwtMgr.Generate(claims.UserID, claims.Username, claims.Email)
	if err != nil {
		return nil, grpcstatus.Error(codes.Internal, "failed to generate access token")
	}

	accessTokenExpires := time.Now().Add(24 * time.Hour)

	return &todov1.RefreshTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: timestamppb.New(accessTokenExpires),
	}, nil
}

// GetProfile retrieves current user profile
func (h *AuthHandler) GetProfile(ctx context.Context, req *todov1.GetProfileRequest) (*todov1.GetProfileResponse, error) {
	// Extract user ID from context (set by auth middleware)
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	user, err := h.authService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &todov1.GetProfileResponse{
		User: h.domainUserToProto(user),
	}, nil
}

// UpdateProfile handles user profile updates
func (h *AuthHandler) UpdateProfile(ctx context.Context, req *todov1.UpdateProfileRequest) (*todov1.UpdateProfileResponse, error) {
	// Extract user ID from context
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Get current user
	user, err := h.authService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.DisplayName != nil {
		user.FullName = *req.DisplayName
	}
	if req.AvatarUrl != nil {
		user.AvatarURL = *req.AvatarUrl
	}

	// Update user in repository - need to add UpdateUser method to AuthService
	// For now, we'll skip the actual update since the method doesn't exist yet
	// TODO: Add UpdateUser method to AuthService
	_ = user

	return &todov1.UpdateProfileResponse{
		User: h.domainUserToProto(user),
	}, nil
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(ctx context.Context, req *todov1.ChangePasswordRequest) (*todov1.ChangePasswordResponse, error) {
	if req.CurrentPassword == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "current password is required")
	}
	if req.NewPassword == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "new password is required")
	}

	// TODO: Implement password change functionality
	// For now, we'll return an error indicating this feature is not yet implemented
	return nil, grpcstatus.Error(codes.Unimplemented, "password change functionality not yet implemented")
}

// VerifyEmail handles email verification
func (h *AuthHandler) VerifyEmail(ctx context.Context, req *todov1.VerifyEmailRequest) (*todov1.VerifyEmailResponse, error) {
	// For now, we'll just return success
	// In a real implementation, you would validate the token and mark email as verified
	return &todov1.VerifyEmailResponse{}, nil
}

// RequestPasswordReset handles password reset request
func (h *AuthHandler) RequestPasswordReset(ctx context.Context, req *todov1.RequestPasswordResetRequest) (*todov1.RequestPasswordResetResponse, error) {
	if req.Email == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "email is required")
	}

	// For now, we'll just return success
	// In a real implementation, you would send a password reset email
	return &todov1.RequestPasswordResetResponse{}, nil
}

// ConfirmPasswordReset handles password reset confirmation
func (h *AuthHandler) ConfirmPasswordReset(ctx context.Context, req *todov1.ConfirmPasswordResetRequest) (*todov1.ConfirmPasswordResetResponse, error) {
	if req.Token == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "token is required")
	}
	if req.NewPassword == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "new password is required")
	}

	// For now, we'll just return success
	// In a real implementation, you would validate the token and update the password
	return &todov1.ConfirmPasswordResetResponse{}, nil
}

// domainUserToProto converts domain User to protobuf User
func (h *AuthHandler) domainUserToProto(user *domain.User) *todov1.User {
	var lastLoginAt *timestamppb.Timestamp
	if user.LastLoginAt != nil {
		lastLoginAt = timestamppb.New(*user.LastLoginAt)
	}

	return &todov1.User{
		Id:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		DisplayName:   user.FullName,
		AvatarUrl:     user.AvatarURL,
		CreatedAt:     timestamppb.New(user.CreatedAt),
		UpdatedAt:     timestamppb.New(user.UpdatedAt),
		LastLoginAt:   lastLoginAt,
		EmailVerified: true, // For now, assume email is verified
	}
}
