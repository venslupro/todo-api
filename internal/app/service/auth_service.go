package service

import (
	"context"
	"fmt"

	"github.com/venslupro/todo-api/internal/domain"
	"github.com/venslupro/todo-api/internal/pkg/auth"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

// AuthService provides authentication and authorization services
type AuthService struct {
	userRepo domain.UserRepository
	jwtMgr   *auth.JWTManager
	password *PasswordManager
}

// PasswordManager wraps password operations
type PasswordManager struct{}

// HashPassword hashes a password
func (p *PasswordManager) HashPassword(password string) (string, error) {
	return auth.HashPassword(password)
}

// CheckPassword checks if a password matches a hash
func (p *PasswordManager) CheckPassword(password, hash string) bool {
	return auth.CheckPassword(password, hash)
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo domain.UserRepository, jwtMgr *auth.JWTManager) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtMgr:   jwtMgr,
		password: &PasswordManager{},
	}
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, email, username, password, fullName string) (*domain.User, error) {
	// Validate input
	if email == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "email is required")
	}
	if username == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "username is required")
	}
	if password == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "password is required")
	}

	// Check if email already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to check email: %v", err))
	}
	if exists {
		return nil, grpcstatus.Error(codes.AlreadyExists, "email already exists")
	}

	// Check if username already exists
	exists, err = s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to check username: %v", err))
	}
	if exists {
		return nil, grpcstatus.Error(codes.AlreadyExists, "username already exists")
	}

	// Hash password
	passwordHash, err := s.password.HashPassword(password)
	if err != nil {
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to hash password: %v", err))
	}

	// Create user
	user := domain.NewUser(email, username, passwordHash)
	if fullName != "" {
		user.FullName = fullName
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to create user: %v", err))
	}

	return user, nil
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", grpcstatus.Error(codes.Unauthenticated, "invalid email or password")
	}

	// Check password
	if !s.password.CheckPassword(password, user.PasswordHash) {
		return nil, "", grpcstatus.Error(codes.Unauthenticated, "invalid email or password")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, "", grpcstatus.Error(codes.PermissionDenied, "user account is inactive")
	}

	// Update last login
	user.UpdateLastLogin()
	if err := s.userRepo.Update(ctx, user); err != nil {
		// Log error but don't fail login
		_ = err
	}

	// Generate token
	token, err := s.jwtMgr.Generate(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, "", grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to generate token: %v", err))
	}

	return user, token, nil
}

// ValidateToken validates a JWT token and returns user information
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*auth.Claims, error) {
	claims, err := s.jwtMgr.Validate(token)
	if err != nil {
		return nil, grpcstatus.Error(codes.Unauthenticated, fmt.Sprintf("invalid token: %v", err))
	}

	// Optionally verify user still exists and is active
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, grpcstatus.Error(codes.Unauthenticated, "user not found")
	}
	if !user.IsActive {
		return nil, grpcstatus.Error(codes.PermissionDenied, "user account is inactive")
	}

	return claims, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	if userID == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "user id is required")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, grpcstatus.Error(codes.NotFound, fmt.Sprintf("user not found: %v", err))
	}

	return user, nil
}
