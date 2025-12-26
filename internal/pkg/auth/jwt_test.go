package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTManager_Generate(t *testing.T) {
	secretKey := "test-secret-key"
	tokenDuration := 24 * time.Hour
	jwtMgr := NewJWTManager(secretKey, tokenDuration)

	tests := []struct {
		name     string
		userID   string
		username string
		email    string
		wantErr  bool
	}{
		{
			name:     "successful token generation",
			userID:   "user-123",
			username: "testuser",
			email:    "test@example.com",
			wantErr:  false,
		},
		{
			name:     "empty user ID",
			userID:   "",
			username: "testuser",
			email:    "test@example.com",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtMgr.Generate(tt.userID, tt.username, tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("Generate() returned empty token")
			}
		})
	}
}

func TestJWTManager_Validate(t *testing.T) {
	secretKey := "test-secret-key"
	tokenDuration := 24 * time.Hour
	jwtMgr := NewJWTManager(secretKey, tokenDuration)

	tests := []struct {
		name       string
		setupToken func() string
		wantValid  bool
		wantUserID string
		wantEmail  string
		wantErr    bool
	}{
		{
			name: "valid token",
			setupToken: func() string {
				token, _ := jwtMgr.Generate("user-123", "testuser", "test@example.com")
				return token
			},
			wantValid:  true,
			wantUserID: "user-123",
			wantEmail:  "test@example.com",
			wantErr:    false,
		},
		{
			name: "invalid token",
			setupToken: func() string {
				return "invalid-token-string"
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "expired token",
			setupToken: func() string {
				// Create a manager with very short duration
				expiredMgr := NewJWTManager(secretKey, -1*time.Hour)
				token, _ := expiredMgr.Generate("user-123", "testuser", "test@example.com")
				return token
			},
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupToken()
			claims, err := jwtMgr.Validate(token)

			if tt.wantErr {
				if err == nil {
					t.Error("Validate() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
				return
			}

			if claims.UserID != tt.wantUserID {
				t.Errorf("Validate() UserID = %v, want %v", claims.UserID, tt.wantUserID)
			}

			if claims.Email != tt.wantEmail {
				t.Errorf("Validate() Email = %v, want %v", claims.Email, tt.wantEmail)
			}
		})
	}
}

func TestJWTManager_GenerateAndValidate(t *testing.T) {
	secretKey := "test-secret-key"
	tokenDuration := 24 * time.Hour
	jwtMgr := NewJWTManager(secretKey, tokenDuration)

	userID := "user-123"
	username := "testuser"
	email := "test@example.com"

	token, err := jwtMgr.Generate(userID, username, email)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	claims, err := jwtMgr.Validate(token)
	if err != nil {
		t.Fatalf("Validate() failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID mismatch: got %v, want %v", claims.UserID, userID)
	}

	if claims.Username != username {
		t.Errorf("Username mismatch: got %v, want %v", claims.Username, username)
	}

	if claims.Email != email {
		t.Errorf("Email mismatch: got %v, want %v", claims.Email, email)
	}

	// Verify token expiration
	if claims.ExpiresAt.Before(time.Now()) {
		t.Error("Token has already expired")
	}

	if claims.IssuedAt.After(time.Now()) {
		t.Error("Token issued in the future")
	}
}

func TestJWTManager_InvalidSecretKey(t *testing.T) {
	jwtMgr1 := NewJWTManager("secret-key-1", 24*time.Hour)
	jwtMgr2 := NewJWTManager("secret-key-2", 24*time.Hour)

	token, err := jwtMgr1.Generate("user-123", "testuser", "test@example.com")
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Try to validate with different secret key
	_, err = jwtMgr2.Validate(token)
	if err == nil {
		t.Error("Validate() should have failed with different secret key")
	}
}

func TestClaims_ImplementsJWTClaims(t *testing.T) {
	var _ jwt.Claims = &Claims{}
}
