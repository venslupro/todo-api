package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "successful password hash",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: "very-long-password-with-special-characters!@#$%^&*()",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && hash == "" {
				t.Error("HashPassword() returned empty hash")
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	password := "password123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	tests := []struct {
		name      string
		password  string
		hash      string
		wantMatch bool
	}{
		{
			name:      "correct password",
			password:  password,
			hash:      hash,
			wantMatch: true,
		},
		{
			name:      "incorrect password",
			password:  "wrongpassword",
			hash:      hash,
			wantMatch: false,
		},
		{
			name:      "empty password",
			password:  "",
			hash:      hash,
			wantMatch: false,
		},
		{
			name:      "empty hash",
			password:  password,
			hash:      "",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := CheckPassword(tt.password, tt.hash)
			if matches != tt.wantMatch {
				t.Errorf("CheckPassword() = %v, want %v", matches, tt.wantMatch)
			}
		})
	}
}

func TestHashPassword_UniqueHashes(t *testing.T) {
	password := "password123"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	// Hashes should be different due to salt
	if hash1 == hash2 {
		t.Error("HashPassword() should generate different hashes for same password")
	}

	// Both hashes should validate correctly
	if !CheckPassword(password, hash1) {
		t.Error("First hash should validate correctly")
	}

	if !CheckPassword(password, hash2) {
		t.Error("Second hash should validate correctly")
	}
}

func TestCheckPassword_InvalidHash(t *testing.T) {
	password := "password123"
	invalidHash := "invalid-hash-string"

	matches := CheckPassword(password, invalidHash)
	if matches {
		t.Error("CheckPassword() should return false for invalid hash")
	}
}

func TestPasswordSecurity(t *testing.T) {
	// Test that different passwords produce different hashes
	password1 := "password123"
	password2 := "password456"

	hash1, err := HashPassword(password1)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	hash2, err := HashPassword(password2)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	if hash1 == hash2 {
		t.Error("Different passwords should produce different hashes")
	}

	// Verify that password1 doesn't work with password2's hash
	if CheckPassword(password1, hash2) {
		t.Error("Password1 should not validate against password2's hash")
	}

	if CheckPassword(password2, hash1) {
		t.Error("Password2 should not validate against password1's hash")
	}
}
