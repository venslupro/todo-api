package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Set up environment variables for testing
	os.Setenv("GRPC_PORT", "50051")
	os.Setenv("HTTP_PORT", "8080")
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("ACCESS_TOKEN_EXPIRY", "1h")
	os.Setenv("REFRESH_TOKEN_EXPIRY", "24h")
	os.Setenv("STORAGE_S3_BUCKET", "test-bucket")
	os.Setenv("STORAGE_S3_REGION", "us-east-1")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify server config
	if config.Server.GRPCPort != 50051 {
		t.Errorf("Server.GRPCPort = %v, want %v", config.Server.GRPCPort, 50051)
	}

	if config.Server.HTTPPort != 8080 {
		t.Errorf("Server.HTTPPort = %v, want %v", config.Server.HTTPPort, 8080)
	}

	if config.Server.Environment != "test" {
		t.Errorf("Server.Environment = %v, want %v", config.Server.Environment, "test")
	}

	// Verify database config
	if config.Database.Host != "localhost" {
		t.Errorf("Database.Host = %v, want %v", config.Database.Host, "localhost")
	}

	if config.Database.Port != 5432 {
		t.Errorf("Database.Port = %v, want %v", config.Database.Port, 5432)
	}

	if config.Database.User != "testuser" {
		t.Errorf("Database.User = %v, want %v", config.Database.User, "testuser")
	}

	if config.Database.Password != "testpass" {
		t.Errorf("Database.Password = %v, want %v", config.Database.Password, "testpass")
	}

	if config.Database.DBName != "testdb" {
		t.Errorf("Database.DBName = %v, want %v", config.Database.DBName, "testdb")
	}

	// Verify auth config
	if config.Auth.JWTSecret != "test-secret" {
		t.Errorf("Auth.JWTSecret = %v, want %v", config.Auth.JWTSecret, "test-secret")
	}

	if config.Auth.AccessTokenExpiry != time.Hour {
		t.Errorf("Auth.AccessTokenExpiry = %v, want %v", config.Auth.AccessTokenExpiry, time.Hour)
	}

	if config.Auth.RefreshTokenExpiry != 24*time.Hour {
		t.Errorf("Auth.RefreshTokenExpiry = %v, want %v", config.Auth.RefreshTokenExpiry, 24*time.Hour)
	}

	// Verify storage config
	if config.Storage.S3Bucket != "test-bucket" {
		t.Errorf("Storage.S3Bucket = %v, want %v", config.Storage.S3Bucket, "test-bucket")
	}

	if config.Storage.S3Region != "us-east-1" {
		t.Errorf("Storage.S3Region = %v, want %v", config.Storage.S3Region, "us-east-1")
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear environment variables to test defaults
	os.Unsetenv("GRPC_PORT")
	os.Unsetenv("HTTP_PORT")
	os.Unsetenv("ENVIRONMENT")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("ACCESS_TOKEN_EXPIRY")
	os.Unsetenv("REFRESH_TOKEN_EXPIRY")
	os.Unsetenv("STORAGE_S3_BUCKET")
	os.Unsetenv("STORAGE_S3_REGION")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify default values
	if config.Server.GRPCPort != 50051 {
		t.Errorf("Server.GRPCPort = %v, want %v", config.Server.GRPCPort, 50051)
	}

	if config.Server.HTTPPort != 8080 {
		t.Errorf("Server.HTTPPort = %v, want %v", config.Server.HTTPPort, 8080)
	}

	if config.Server.Environment != "development" {
		t.Errorf("Server.Environment = %v, want %v", config.Server.Environment, "development")
	}

	if config.Database.Host != "localhost" {
		t.Errorf("Database.Host = %v, want %v", config.Database.Host, "localhost")
	}

	if config.Database.Port != 5432 {
		t.Errorf("Database.Port = %v, want %v", config.Database.Port, 5432)
	}

	if config.Database.User != "postgres" {
		t.Errorf("Database.User = %v, want %v", config.Database.User, "postgres")
	}

	if config.Database.Password != "postgres" {
		t.Errorf("Database.Password = %v, want %v", config.Database.Password, "postgres")
	}

	if config.Database.DBName != "todo_db" {
		t.Errorf("Database.DBName = %v, want %v", config.Database.DBName, "todo_db")
	}

	if config.Database.SSLMode != "disable" {
		t.Errorf("Database.SSLMode = %v, want %v", config.Database.SSLMode, "disable")
	}

	if config.Redis.Host != "localhost" {
		t.Errorf("Redis.Host = %v, want %v", config.Redis.Host, "localhost")
	}

	if config.Redis.Port != 6379 {
		t.Errorf("Redis.Port = %v, want %v", config.Redis.Port, 6379)
	}

	if config.Auth.JWTSecret != "your-secret-key-change-in-production" {
		t.Errorf("Auth.JWTSecret = %v, want %v", config.Auth.JWTSecret, "your-secret-key-change-in-production")
	}

	if config.Auth.AccessTokenExpiry != 15*time.Minute {
		t.Errorf("Auth.AccessTokenExpiry = %v, want %v", config.Auth.AccessTokenExpiry, 15*time.Minute)
	}

	if config.Auth.RefreshTokenExpiry != 7*24*time.Hour {
		t.Errorf("Auth.RefreshTokenExpiry = %v, want %v", config.Auth.RefreshTokenExpiry, 7*24*time.Hour)
	}

	if config.Storage.S3Bucket != "" {
		t.Errorf("Storage.S3Bucket = %v, want %v", config.Storage.S3Bucket, "")
	}

	if config.Storage.S3Region != "us-east-1" {
		t.Errorf("Storage.S3Region = %v, want %v", config.Storage.S3Region, "us-east-1")
	}
}

func TestLoadConfig_InvalidDuration(t *testing.T) {
	// Set invalid duration format
	os.Setenv("ACCESS_TOKEN_EXPIRY", "invalid-duration")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Should use default value when duration is invalid
	if config.Auth.AccessTokenExpiry != 15*time.Minute {
		t.Errorf("Auth.AccessTokenExpiry = %v, want %v", config.Auth.AccessTokenExpiry, 15*time.Minute)
	}
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "valid config",
			setup: func() {
				os.Setenv("GRPC_PORT", "50051")
				os.Setenv("HTTP_PORT", "8080")
				os.Setenv("ENVIRONMENT", "test")
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_PORT", "5432")
				os.Setenv("DB_USER", "testuser")
				os.Setenv("DB_PASSWORD", "testpass")
				os.Setenv("DB_NAME", "testdb")
				os.Setenv("JWT_SECRET", "test-secret")
			},
			wantErr: false,
		},
		{
			name: "invalid GRPC port",
			setup: func() {
				os.Setenv("GRPC_PORT", "-1")
			},
			wantErr: false, // Should use default value
		},
		{
			name: "invalid HTTP port",
			setup: func() {
				os.Setenv("HTTP_PORT", "70000")
			},
			wantErr: false, // Should use default value
		},
		{
			name: "invalid database port",
			setup: func() {
				os.Setenv("DB_PORT", "99999")
			},
			wantErr: false, // Should use default value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset environment
			os.Unsetenv("GRPC_PORT")
			os.Unsetenv("HTTP_PORT")
			os.Unsetenv("DB_PORT")

			// Setup test environment
			tt.setup()

			_, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue string
		want         string
	}{
		{
			name:         "environment variable exists",
			key:          "TEST_KEY",
			value:        "test-value",
			defaultValue: "default",
			want:         "test-value",
		},
		{
			name:         "environment variable does not exist",
			key:          "NON_EXISTENT_KEY",
			value:        "",
			defaultValue: "default",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.want {
				t.Errorf("getEnv(%v, %v) = %v, want %v", tt.key, tt.defaultValue, result, tt.want)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue int
		want         int
	}{
		{
			name:         "valid integer",
			key:          "TEST_INT",
			value:        "123",
			defaultValue: 0,
			want:         123,
		},
		{
			name:         "invalid integer",
			key:          "TEST_INT",
			value:        "not-a-number",
			defaultValue: 999,
			want:         999,
		},
		{
			name:         "variable does not exist",
			key:          "NON_EXISTENT_INT",
			value:        "",
			defaultValue: 999,
			want:         999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnvInt(tt.key, tt.defaultValue)
			if result != tt.want {
				t.Errorf("getEnvInt(%v, %v) = %v, want %v", tt.key, tt.defaultValue, result, tt.want)
			}
		})
	}
}

func TestGetEnvDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue time.Duration
		want         time.Duration
	}{
		{
			name:         "valid duration",
			key:          "TEST_DURATION",
			value:        "1h30m",
			defaultValue: 0,
			want:         90 * time.Minute,
		},
		{
			name:         "invalid duration",
			key:          "TEST_DURATION",
			value:        "invalid-duration",
			defaultValue: time.Hour,
			want:         time.Hour,
		},
		{
			name:         "variable does not exist",
			key:          "NON_EXISTENT_DURATION",
			value:        "",
			defaultValue: time.Hour,
			want:         time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnvDuration(tt.key, tt.defaultValue)
			if result != tt.want {
				t.Errorf("getEnvDuration(%v, %v) = %v, want %v", tt.key, tt.defaultValue, result, tt.want)
			}
		})
	}
}
