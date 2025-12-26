# TODO API Development Guide

> 📖 **Back to [README.md](../README.md)** | **Previous: [Architecture Overview](ARCHITECTURE.md)**

This guide provides comprehensive information for developers working on the TODO API project, including setup instructions, coding standards, testing guidelines, and contribution workflows.

## 🚀 Development Setup

### Prerequisites

- **Go 1.25.5+** - [Download](https://golang.org/dl/)
- **PostgreSQL 13+** - [Download](https://www.postgresql.org/download/)
- **Redis 6+** - [Download](https://redis.io/download)
- **Git** - Version control
- **Docker** (optional) - Containerization
- **Protobuf Compiler** - Protocol Buffer compilation

### Quick Start

1. **Clone the Repository**
```bash
git clone https://github.com/venslupro/todo-api.git
cd todo-api
```

2. **Set Up Environment**
```bash
# Copy environment template
cp .env.example .env

# Install Go dependencies
go mod download

# Install development tools
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
```

3. **Start Development Services**
```bash
# Start PostgreSQL and Redis
docker run -d --name postgres \
  -e POSTGRES_DB=todo_api \
  -e POSTGRES_USER=todo_user \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  postgres:13

docker run -d --name redis -p 6379:6379 redis:6-alpine
```

4. **Run Database Migrations**
```bash
go run cmd/server/migrate.go
```

5. **Start Development Server**
```bash
go run cmd/server/main.go
```

## 🏗️ Project Structure

```
todo-api/
├── api/                    # API definitions
│   ├── proto/             # Protocol Buffer definitions
│   └── gen/               # Generated code
├── cmd/                   # Application entry points
│   └── server/            # Main server
├── internal/              # Private application code
│   ├── app/               # Application layer
│   │   ├── handlers/      # HTTP/gRPC handlers
│   │   ├── routes/        # Route definitions
│   │   └── service/       # Business logic services
│   ├── config/            # Configuration management
│   ├── domain/            # Domain models and interfaces
│   └── infrastructure/    # External integrations
│       ├── database/      # Database layer
│       └── redis/         # Redis client
├── pkg/                   # Public libraries
└── deployment/            # Deployment configurations
```

## 📝 Coding Standards

### Go Code Style

Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and these project-specific guidelines:

#### Naming Conventions

- **Packages**: lowercase, single word
- **Interfaces**: `-er` suffix (e.g., `Repository`, `Handler`)
- **Variables**: camelCase
- **Constants**: UPPER_SNAKE_CASE
- **Exported functions**: PascalCase

#### Code Organization

```go
// File structure example
package todo

import (
    // Standard library
    "context"
    "fmt"
    
    // Third-party libraries
    "github.com/google/uuid"
    
    // Internal packages
    "github.com/venslupro/todo-api/internal/domain"
)

// Type definitions
type Service struct {
    repo domain.TodoRepository
}

// Constructor
func NewService(repo domain.TodoRepository) *Service {
    return &Service{repo: repo}
}

// Public methods
func (s *Service) CreateTodo(ctx context.Context, req *CreateRequest) (*Todo, error) {
    // Implementation
}

// Private methods
func (s *Service) validateTodo(todo *Todo) error {
    // Implementation
}
```

### Error Handling

Use Go's error handling patterns consistently:

```go
// Good: Wrapping errors with context
func (s *Service) GetTodo(ctx context.Context, id string) (*Todo, error) {
    todo, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get todo %s: %w", id, err)
    }
    
    if todo == nil {
        return nil, domain.ErrTodoNotFound
    }
    
    return todo, nil
}

// Good: Using custom error types
var (
    ErrTodoNotFound = errors.New("todo not found")
    ErrInvalidInput = errors.New("invalid input")
)
```

### Logging

Use structured logging with context:

```go
import "github.com/sirupsen/logrus"

func (s *Service) ProcessTodo(ctx context.Context, todoID string) error {
    logger := logrus.WithFields(logrus.Fields{
        "todo_id": todoID,
        "user_id": ctx.Value("user_id"),
    })
    
    logger.Info("Processing todo")
    
    // Processing logic
    
    logger.Info("Todo processed successfully")
    return nil
}
```

## 🔧 Development Workflow

### Code Generation

#### Protocol Buffer Compilation

```bash
# Generate Go code from .proto files
buf generate

# Or manually with protoc
protoc --go_out=. --go-grpc_out=. \
  --grpc-gateway_out=. \
  --openapiv2_out=. \
  api/proto/todo/v1/*.proto
```

#### Database Migrations

```bash
# Create new migration
migrate create -ext sql -dir internal/infrastructure/database/migrations add_user_preferences

# Run migrations
go run cmd/server/migrate.go

# Rollback last migration
go run cmd/server/migrate.go -rollback
```

### Testing

#### Running Tests

```bash
# Run all tests
go test ./... -v

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/app/service/... -v

# Run tests with race detector
go test -race ./...
```

#### Writing Tests

Follow the table-driven test pattern:

```go
func TestTodoService_Create(t *testing.T) {
    tests := []struct {
        name        string
        input       *CreateRequest
        setup       func(*mockRepository)
        want        *Todo
        wantErr     bool
        errContains string
    }{
        {
            name: "successful creation",
            input: &CreateRequest{
                Title: "Test Todo",
                Description: "Test description",
            },
            setup: func(mock *mockRepository) {
                mock.On("Create", mock.Anything, mock.Anything).
                    Return(&domain.Todo{ID: "123"}, nil)
            },
            want: &Todo{ID: "123", Title: "Test Todo"},
            wantErr: false,
        },
        {
            name: "empty title",
            input: &CreateRequest{
                Title: "",
                Description: "Test description",
            },
            wantErr: true,
            errContains: "title is required",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockRepo := new(mockRepository)
            if tt.setup != nil {
                tt.setup(mockRepo)
            }
            
            service := NewTodoService(mockRepo)
            
            // Execute
            got, err := service.Create(context.Background(), tt.input)
            
            // Verify
            if (err != nil) != tt.wantErr {
                t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if tt.wantErr {
                if !strings.Contains(err.Error(), tt.errContains) {
                    t.Errorf("Create() error = %v, should contain %v", err, tt.errContains)
                }
                return
            }
            
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Create() = %v, want %v", got, tt.want)
            }
            
            mockRepo.AssertExpectations(t)
        })
    }
}
```

### Debugging

#### Using Delve Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug the application
dlv debug cmd/server/main.go

# Set breakpoints and run
(dlv) break internal/app/service/todo_service.go:45
(dlv) continue
```

#### Logging for Debugging

```go
// Use debug-level logging for troubleshooting
logger := logrus.WithFields(logrus.Fields{
    "todo_id": todoID,
    "user_id": userID,
})

logger.Debug("Starting todo processing")
// ... processing logic
logger.Debug("Todo processing completed")
```

## 🧪 Testing Strategy

### Unit Tests

- **Coverage**: Aim for 80%+ test coverage
- **Mocking**: Use interfaces for testability
- **Isolation**: Test one component at a time
- **Table-driven**: Use table-driven tests for multiple scenarios

### Integration Tests

```go
func TestTodoIntegration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()
    
    // Create service with real dependencies
    repo := database.NewTodoRepository(db)
    service := NewTodoService(repo)
    
    // Test real interactions
    todo, err := service.Create(context.Background(), &CreateRequest{
        Title: "Integration Test",
    })
    
    assert.NoError(t, err)
    assert.NotNil(t, todo)
    assert.Equal(t, "Integration Test", todo.Title)
}
```

### End-to-End Tests

```bash
# Run E2E tests with docker-compose
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

## 🔄 Continuous Integration

### GitHub Actions

The project includes CI/CD workflows in `.github/workflows/ci-cd.yml`:

```yaml
name: CI/CD

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:13
        env:
          POSTGRES_PASSWORD: password
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:6-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.25'
    
    - name: Run tests
      run: |
        go test ./... -v -race -coverprofile=coverage.out
        go tool cover -func=coverage.out
```

## 📚 API Development

### Adding New API Endpoints

1. **Define Protocol Buffer**
```protobuf
service TodoService {
    rpc CreateTodo(CreateTodoRequest) returns (Todo) {
        option (google.api.http) = {
            post: "/v1/todos"
            body: "*"
        };
    }
}

message CreateTodoRequest {
    string title = 1;
    string description = 2;
}
```

2. **Generate Code**
```bash
buf generate
```

3. **Implement Service**
```go
type todoServiceServer struct {
    service app.TodoService
}

func (s *todoServiceServer) CreateTodo(ctx context.Context, req *todov1.CreateTodoRequest) (*todov1.Todo, error) {
    // Convert request to domain model
    todo, err := s.service.Create(ctx, &app.CreateRequest{
        Title: req.Title,
        Description: req.Description,
    })
    
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to create todo: %v", err)
    }
    
    // Convert domain model to response
    return &todov1.Todo{
        Id:          todo.ID,
        Title:       todo.Title,
        Description: todo.Description,
    }, nil
}
```

### Database Schema Changes

1. **Create Migration**
```sql
-- migrations/003_add_priority_column.up.sql
ALTER TABLE todos ADD COLUMN priority INTEGER NOT NULL DEFAULT 0;

-- migrations/003_add_priority_column.down.sql
ALTER TABLE todos DROP COLUMN priority;
```

2. **Update Domain Model**
```go
type Todo struct {
    ID          string    `db:"id"`
    Title       string    `db:"title"`
    Description string    `db:"description"`
    Priority    int       `db:"priority"`
    CreatedAt   time.Time `db:"created_at"`
    UpdatedAt   time.Time `db:"updated_at"`
}
```

## 🐛 Common Issues and Solutions

### Database Connection Issues

**Problem**: Database connection failures
**Solution**: Check connection string and network connectivity

```go
// Test database connection
func TestDBConnection(t *testing.T) {
    db, err := sql.Open("postgres", connString)
    if err != nil {
        t.Fatalf("Failed to connect: %v", err)
    }
    defer db.Close()
    
    if err := db.Ping(); err != nil {
        t.Fatalf("Failed to ping: %v", err)
    }
}
```

### Protocol Buffer Compilation Errors

**Problem**: `protoc` command fails
**Solution**: Ensure all dependencies are installed

```bash
# Install protoc
brew install protobuf  # macOS
sudo apt install protobuf-compiler  # Ubuntu

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Race Conditions

**Problem**: Intermittent test failures
**Solution**: Use race detector and proper synchronization

```bash
# Run tests with race detector
go test -race ./...

# Use mutexes for shared resources
var mu sync.Mutex
mu.Lock()
defer mu.Unlock()
// critical section
```

## 🤝 Contributing

### Branch Strategy

- `main` - Production-ready code
- `develop` - Development branch
- `feature/` - Feature branches
- `bugfix/` - Bug fix branches
- `release/` - Release preparation

### Pull Request Process

1. **Create Feature Branch**
```bash
git checkout -b feature/add-user-preferences
```

2. **Make Changes and Test**
```bash
# Run tests before committing
go test ./... -v

# Format code
go fmt ./...

# Check for lint issues
golangci-lint run
```

3. **Commit Changes**
```bash
git add .
git commit -m "feat: add user preferences API"
```

4. **Push and Create PR**
```bash
git push origin feature/add-user-preferences
# Create PR on GitHub
```

### Code Review Guidelines

- **Review your own code** before requesting review
- **Keep PRs small and focused**
- **Include tests** for new functionality
- **Update documentation** as needed
- **Follow the established patterns**

This development guide provides comprehensive information for working effectively on the TODO API project, ensuring code quality, maintainability, and team collaboration.