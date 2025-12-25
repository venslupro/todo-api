package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/venslupro/todo-api/internal/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuthInterceptor(t *testing.T) {
	jwtMgr := auth.NewJWTManager("test-secret", 24*time.Hour)
	interceptor := AuthInterceptor(jwtMgr)

	tests := []struct {
		name       string
		fullMethod string
		setupCtx   func() context.Context
		wantErr    bool
		errorCode  codes.Code
	}{
		{
			name:       "skip auth for register method",
			fullMethod: "/todo.v1.AuthService/Register",
			setupCtx:   func() context.Context { return context.Background() },
			wantErr:    false,
		},
		{
			name:       "skip auth for login method",
			fullMethod: "/todo.v1.AuthService/Login",
			setupCtx:   func() context.Context { return context.Background() },
			wantErr:    false,
		},
		{
			name:       "skip auth for health check",
			fullMethod: "/todo.v1.SystemService/HealthCheck",
			setupCtx:   func() context.Context { return context.Background() },
			wantErr:    false,
		},
		{
			name:       "missing metadata",
			fullMethod: "/todo.v1.TodoService/CreateTodo",
			setupCtx:   func() context.Context { return context.Background() },
			wantErr:    true,
			errorCode:  codes.Unauthenticated,
		},
		{
			name:       "missing authorization header",
			fullMethod: "/todo.v1.TodoService/CreateTodo",
			setupCtx: func() context.Context {
				md := metadata.New(map[string]string{})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantErr:   true,
			errorCode: codes.Unauthenticated,
		},
		{
			name:       "invalid authorization format",
			fullMethod: "/todo.v1.TodoService/CreateTodo",
			setupCtx: func() context.Context {
				md := metadata.New(map[string]string{
					"authorization": "InvalidToken",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantErr:   true,
			errorCode: codes.Unauthenticated,
		},
		{
			name:       "invalid token",
			fullMethod: "/todo.v1.TodoService/CreateTodo",
			setupCtx: func() context.Context {
				md := metadata.New(map[string]string{
					"authorization": "Bearer invalid-token",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantErr:   true,
			errorCode: codes.Unauthenticated,
		},
		{
			name:       "valid token",
			fullMethod: "/todo.v1.TodoService/CreateTodo",
			setupCtx: func() context.Context {
				token, _ := jwtMgr.Generate("user-123", "testuser", "test@example.com")
				md := metadata.New(map[string]string{
					"authorization": "Bearer " + token,
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			info := &grpc.UnaryServerInfo{
				FullMethod: tt.fullMethod,
			}

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return "success", nil
			}

			_, err := interceptor(ctx, "test-request", info, handler)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
					return
				}

				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("error is not a gRPC status error: %v", err)
					return
				}

				if st.Code() != tt.errorCode {
					t.Errorf("error code = %v, want %v", st.Code(), tt.errorCode)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name     string
		setupCtx func() context.Context
		wantID   string
		wantErr  bool
	}{
		{
			name: "valid user ID",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), UserIDKey, "user-123")
			},
			wantID:  "user-123",
			wantErr: false,
		},
		{
			name: "empty user ID",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), UserIDKey, "")
			},
			wantErr: true,
		},
		{
			name: "missing user ID",
			setupCtx: func() context.Context {
				return context.Background()
			},
			wantErr: true,
		},
		{
			name: "wrong type",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), UserIDKey, 123)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			userID, err := GetUserIDFromContext(ctx)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if userID != tt.wantID {
					t.Errorf("user ID = %v, want %v", userID, tt.wantID)
				}
			}
		})
	}
}

func TestGetUsernameFromContext(t *testing.T) {
	tests := []struct {
		name       string
		setupCtx   func() context.Context
		wantResult string
	}{
		{
			name: "valid username",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), UsernameKey, "testuser")
			},
			wantResult: "testuser",
		},
		{
			name: "empty username",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), UsernameKey, "")
			},
			wantResult: "",
		},
		{
			name: "missing username",
			setupCtx: func() context.Context {
				return context.Background()
			},
			wantResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			username := GetUsernameFromContext(ctx)

			if username != tt.wantResult {
				t.Errorf("username = %v, want %v", username, tt.wantResult)
			}
		})
	}
}

func TestGetEmailFromContext(t *testing.T) {
	tests := []struct {
		name       string
		setupCtx   func() context.Context
		wantResult string
	}{
		{
			name: "valid email",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), EmailKey, "test@example.com")
			},
			wantResult: "test@example.com",
		},
		{
			name: "empty email",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), EmailKey, "")
			},
			wantResult: "",
		},
		{
			name: "missing email",
			setupCtx: func() context.Context {
				return context.Background()
			},
			wantResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			email := GetEmailFromContext(ctx)

			if email != tt.wantResult {
				t.Errorf("email = %v, want %v", email, tt.wantResult)
			}
		})
	}
}

func TestShouldSkipAuth(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		wantResult bool
	}{
		{
			name:       "register method",
			method:     "/todo.v1.AuthService/Register",
			wantResult: true,
		},
		{
			name:       "login method",
			method:     "/todo.v1.AuthService/Login",
			wantResult: true,
		},
		{
			name:       "health check method",
			method:     "/todo.v1.SystemService/HealthCheck",
			wantResult: true,
		},
		{
			name:       "other method",
			method:     "/todo.v1.TodoService/CreateTodo",
			wantResult: false,
		},
		{
			name:       "unknown method",
			method:     "/unknown.Service/Method",
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSkipAuth(tt.method)
			if result != tt.wantResult {
				t.Errorf("shouldSkipAuth(%v) = %v, want %v", tt.method, result, tt.wantResult)
			}
		})
	}
}
