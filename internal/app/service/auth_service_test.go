package service

import (
	"context"
	"testing"
	"time"

	"github.com/venslupro/todo-api/internal/domain"
	"github.com/venslupro/todo-api/internal/pkg/auth"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

// MockUserRepository is a mock implementation of UserRepository for testing
type MockUserRepository struct {
	users map[string]*domain.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*domain.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, &NotFoundError{ID: id}
	}
	return user, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, &NotFoundError{ID: email}
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, &NotFoundError{ID: username}
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	if _, ok := m.users[user.ID]; !ok {
		return &NotFoundError{ID: user.ID}
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	if _, ok := m.users[id]; !ok {
		return &NotFoundError{ID: id}
	}
	delete(m.users, id)
	return nil
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	for _, user := range m.users {
		if user.Email == email {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	for _, user := range m.users {
		if user.Username == username {
			return true, nil
		}
	}
	return false, nil
}

func TestAuthService_Register(t *testing.T) {
	ctx := context.Background()
	userRepo := NewMockUserRepository()
	jwtMgr := auth.NewJWTManager("test-secret", 24*time.Hour)
	authService := NewAuthService(userRepo, jwtMgr)

	tests := []struct {
		name        string
		email       string
		username    string
		password    string
		fullName    string
		expectError bool
		errorCode   codes.Code
	}{
		{
			name:        "successful registration",
			email:       "test@example.com",
			username:    "testuser",
			password:    "password123",
			fullName:    "Test User",
			expectError: false,
		},
		{
			name:        "duplicate email",
			email:       "test@example.com",
			username:    "testuser2",
			password:    "password123",
			fullName:    "Test User 2",
			expectError: true,
			errorCode:   codes.AlreadyExists,
		},
		{
			name:        "duplicate username",
			email:       "test2@example.com",
			username:    "testuser",
			password:    "password123",
			fullName:    "Test User 2",
			expectError: true,
			errorCode:   codes.AlreadyExists,
		},
		{
			name:        "empty email",
			email:       "",
			username:    "testuser3",
			password:    "password123",
			fullName:    "Test User 3",
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := authService.Register(ctx, tt.email, tt.username, tt.password, tt.fullName)

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
				if user.Email != tt.email {
					t.Errorf("expected email %s, got %s", tt.email, user.Email)
				}
				if user.Username != tt.username {
					t.Errorf("expected username %s, got %s", tt.username, user.Username)
				}
				if user.FullName != tt.fullName {
					t.Errorf("expected full name %s, got %s", tt.fullName, user.FullName)
				}
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()
	userRepo := NewMockUserRepository()
	jwtMgr := auth.NewJWTManager("test-secret", 24*time.Hour)
	authService := NewAuthService(userRepo, jwtMgr)

	// First register a user
	user, err := authService.Register(ctx, "test@example.com", "testuser", "password123", "Test User")
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	tests := []struct {
		name        string
		email       string
		password    string
		expectError bool
		errorCode   codes.Code
	}{
		{
			name:        "successful login",
			email:       "test@example.com",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "wrong password",
			email:       "test@example.com",
			password:    "wrongpassword",
			expectError: true,
			errorCode:   codes.Unauthenticated,
		},
		{
			name:        "non-existent email",
			email:       "nonexistent@example.com",
			password:    "password123",
			expectError: true,
			errorCode:   codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, token, err := authService.Login(ctx, tt.email, tt.password)

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
				if token == "" {
					t.Error("expected token but got empty string")
				}

				// Validate the token
				claims, err := jwtMgr.Validate(token)
				if err != nil {
					t.Fatalf("failed to validate token: %v", err)
				}
				if claims.UserID != user.ID {
					t.Errorf("expected user ID %s, got %s", user.ID, claims.UserID)
				}
			}
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	ctx := context.Background()
	userRepo := NewMockUserRepository()
	jwtMgr := auth.NewJWTManager("test-secret", 24*time.Hour)
	authService := NewAuthService(userRepo, jwtMgr)

	// Register a user
	user, err := authService.Register(ctx, "test@example.com", "testuser", "password123", "Test User")
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	// Generate a valid token
	token, err := jwtMgr.Generate(user.ID, user.Username, user.Email)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	tests := []struct {
		name        string
		token       string
		expectError bool
		errorCode   codes.Code
	}{
		{
			name:        "valid token",
			token:       token,
			expectError: false,
		},
		{
			name:        "invalid token",
			token:       "invalid-token",
			expectError: true,
			errorCode:   codes.Unauthenticated,
		},
		{
			name:        "empty token",
			token:       "",
			expectError: true,
			errorCode:   codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := authService.ValidateToken(ctx, tt.token)

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
				if claims.UserID != user.ID {
					t.Errorf("expected user ID %s, got %s", user.ID, claims.UserID)
				}
				if claims.Email != user.Email {
					t.Errorf("expected email %s, got %s", user.Email, claims.Email)
				}
			}
		})
	}
}

func TestAuthService_GetUserByID(t *testing.T) {
	ctx := context.Background()
	userRepo := NewMockUserRepository()
	jwtMgr := auth.NewJWTManager("test-secret", 24*time.Hour)
	authService := NewAuthService(userRepo, jwtMgr)

	// Register a user
	user, err := authService.Register(ctx, "test@example.com", "testuser", "password123", "Test User")
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	tests := []struct {
		name        string
		userID      string
		expectError bool
		errorCode   codes.Code
	}{
		{
			name:        "valid user ID",
			userID:      user.ID,
			expectError: false,
		},
		{
			name:        "empty user ID",
			userID:      "",
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name:        "non-existent user ID",
			userID:      "non-existent-id",
			expectError: true,
			errorCode:   codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrievedUser, err := authService.GetUserByID(ctx, tt.userID)

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
				if retrievedUser.ID != user.ID {
					t.Errorf("expected user ID %s, got %s", user.ID, retrievedUser.ID)
				}
			}
		})
	}
}
