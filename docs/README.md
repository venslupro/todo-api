# TODO API - Real-time Collaborative TODO System

A scalable and well-designed real-time collaborative TODO list API application built with Go, gRPC, and gRPC-Gateway. The system supports user management, team collaboration, real-time updates, and comprehensive TODO management features.

## Features

### Core Features
- ✅ **TODOs CRUD Operations**
  - Create, Read, Update, Delete TODO items
  - Each TODO has:
    - Unique ID
    - Title
    - Description
    - Due Date
    - Status (Not Started, In Progress, Completed, Cancelled)
    - Priority (Low, Medium, High, Urgent)
    - Tags
    - Media attachments
    - Assignment to users
    - Parent-child relationships (subtasks)
    - Position for manual ordering

- ✅ **Filtering**
  - Filter by status
  - Filter by priority
  - Filter by due date range
  - Filter by tags
  - Filter by assignee
  - Filter by parent TODO
  - Full-text search

- ✅ **Sorting**
  - Sort by due date
  - Sort by status
  - Sort by priority
  - Sort by title
  - Sort by created/updated date
  - Multiple sort criteria support

- ✅ **Pagination**
  - Page-based pagination
  - Configurable page size (max 100 items per page)

### Nice-to-have Features
- ✅ Additional attributes (priority, tags, media attachments)
- ✅ User assignment
- ✅ Subtasks (parent-child relationships)
- ✅ Bulk operations (bulk status update, bulk delete)
- ✅ Complete/Reopen TODO operations
- ✅ Move TODO (change position or parent)

## Architecture

The application follows a clean architecture pattern with clear separation of concerns:

```
├── api/                    # Protocol buffer definitions and generated code
│   ├── proto/             # .proto files
│   └── gen/               # Generated Go code from protos
├── cmd/                   # Application entry points
│   └── server/           # Main server application
├── internal/             # Internal application code
│   ├── app/              # Application layer
│   │   ├── handlers/     # gRPC handlers
│   │   └── service/      # Business logic services
│   ├── config/           # Configuration management
│   ├── domain/           # Domain models and interfaces
│   └── infrastructure/    # Infrastructure implementations
│       └── database/     # Database repository and migrations
└── pkg/                  # Public libraries (if any)
```

### Design Principles
- **SOLID Principles**: Single Responsibility, Open/Closed, Liskov Substitution, Interface Segregation, Dependency Inversion
- **Clean Architecture**: Separation of concerns with domain, application, and infrastructure layers
- **Repository Pattern**: Abstract data access through interfaces
- **Dependency Injection**: Services and handlers receive dependencies through constructors

## Technology Stack

- **Language**: Go 1.25.5
- **Framework**: gRPC with gRPC-Gateway for REST API
- **Database**: PostgreSQL
- **Protocol Buffers**: For API definition and code generation
- **Dependencies**:
  - `github.com/grpc-ecosystem/grpc-gateway/v2` - gRPC to REST gateway
  - `github.com/lib/pq` - PostgreSQL driver
  - `github.com/google/uuid` - UUID generation

## Prerequisites

- Go 1.25.5 or later
- PostgreSQL 12 or later
- Protocol Buffer compiler (protoc)
- buf CLI (for protocol buffer management)

## Setup

### 1. Clone the Repository

```bash
git clone https://github.com/venslupro/todo-api.git
cd todo-api
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Set Up Database

Create a PostgreSQL database:

```bash
createdb todo_db
```

Or using psql:

```sql
CREATE DATABASE todo_db;
```

### 4. Run Migrations

Run the database migrations to create the necessary tables:

```bash
psql -d todo_db -f internal/infrastructure/database/migrations/001_create_todos_table.up.sql
```

### 5. Configure Environment Variables

Create a `.env` file or set the following environment variables:

```bash
# Server Configuration
GRPC_PORT=50051
HTTP_PORT=8080
ENVIRONMENT=development

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=todo_db
DB_SSLMODE=disable
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE_CONNS=5

# Authentication (for future use)
JWT_SECRET=your-secret-key-change-in-production
ACCESS_TOKEN_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=168h

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### 6. Run the Application

```bash
go run cmd/server/main.go
```

The server will start:
- gRPC server on port `50051`
- HTTP/REST server (via gRPC-Gateway) on port `8080`

## API Documentation

### gRPC API

The gRPC API is defined in `api/proto/todo/v1/todo_service.proto`. You can use any gRPC client to interact with the service.

### REST API (via gRPC-Gateway)

The REST API is automatically generated from the gRPC service definitions. All endpoints are prefixed with `/v1/todos`.

#### Endpoints

**Create TODO**
```http
POST /v1/todos
Content-Type: application/json

{
  "title": "Complete project",
  "description": "Finish the TODO API project",
  "status": "STATUS_NOT_STARTED",
  "priority": "PRIORITY_HIGH",
  "due_date": "2025-12-18T23:59:59Z",
  "tags": ["work", "urgent"]
}
```

**Get TODO**
```http
GET /v1/todos/{id}
```

**Update TODO**
```http
PUT /v1/todos/{id}
Content-Type: application/json

{
  "title": "Updated title",
  "status": "STATUS_IN_PROGRESS"
}
```

**Delete TODO**
```http
DELETE /v1/todos/{id}
```

**List TODOs**
```http
GET /v1/todos?status=STATUS_NOT_STARTED&page=1&page_size=20
```

Query Parameters:
- `ids`: Filter by specific IDs (comma-separated)
- `user_id`: Filter by user ID
- `statuses`: Filter by statuses (comma-separated)
- `priorities`: Filter by priorities (comma-separated)
- `tags`: Filter by tags (comma-separated)
- `assigned_to`: Filter by assignee
- `parent_id`: Filter by parent TODO
- `search_query`: Full-text search
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 20, max: 100)

**Bulk Update Status**
```http
POST /v1/todos/bulk/status
Content-Type: application/json

{
  "ids": ["id1", "id2", "id3"],
  "status": "STATUS_COMPLETED"
}
```

**Bulk Delete**
```http
POST /v1/todos/bulk/delete
Content-Type: application/json

{
  "ids": ["id1", "id2", "id3"]
}
```

**Complete TODO**
```http
POST /v1/todos/{id}/complete
```

**Reopen TODO**
```http
POST /v1/todos/{id}/reopen
```

**Move TODO**
```http
POST /v1/todos/{id}/move
Content-Type: application/json

{
  "parent_id": "parent-todo-id",
  "position": 5
}
```

### Swagger Documentation

Swagger/OpenAPI documentation is generated from the protocol buffer definitions and available at:
- Swagger JSON: `api/gen/todo/v1/todo.swagger.json`

You can view the Swagger documentation using tools like Swagger UI or import it into Postman.

## Testing

### Unit Tests

Run unit tests:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

### Integration Tests

Integration tests require a running PostgreSQL instance. Set up a test database and run:

```bash
go test -tags=integration ./...
```

## Development

### Code Generation

If you modify the `.proto` files, regenerate the Go code:

```bash
cd api
buf generate
```

### Database Migrations

To create a new migration:

1. Create `up` and `down` SQL files in `internal/infrastructure/database/migrations/`
2. Name them with a sequential number, e.g., `002_add_index.up.sql` and `002_add_index.down.sql`

## Project Structure

```
todo-api/
├── api/                          # API definitions
│   ├── proto/                    # Protocol buffer definitions
│   │   ├── common/v1/           # Common types (enums, errors, pagination)
│   │   └── todo/v1/             # TODO service definitions
│   └── gen/                      # Generated code
├── build/                        # Build configurations
│   └── package/                 # Docker files
├── cmd/                          # Application entry points
│   └── server/                  # Main server
├── deployments/                  # Deployment configurations
│   ├── docker/                  # Docker compose files
│   └── kubernetes/              # Kubernetes manifests
├── internal/                     # Internal application code
│   ├── app/                     # Application layer
│   │   ├── handlers/            # gRPC handlers
│   │   └── service/             # Business logic
│   ├── config/                  # Configuration
│   ├── domain/                  # Domain models
│   └── infrastructure/         # Infrastructure
│       └── database/           # Database layer
│           └── migrations/     # Database migrations
└── scripts/                     # Utility scripts
```

## Error Handling

The API uses gRPC status codes for error handling:
- `INVALID_ARGUMENT`: Invalid input parameters
- `NOT_FOUND`: Resource not found
- `UNAUTHENTICATED`: Authentication required
- `PERMISSION_DENIED`: Insufficient permissions
- `INTERNAL`: Internal server error

## Future Enhancements

- [ ] User authentication and authorization
- [ ] JWT token-based authentication
- [ ] Team features and collaboration
- [ ] Real-time updates via WebSocket
- [ ] Media upload and storage
- [ ] Activity feed and audit logs
- [ ] Advanced search with Elasticsearch
- [ ] Caching with Redis
- [ ] Rate limiting
- [ ] API versioning
- [ ] Comprehensive test coverage
- [ ] CI/CD pipeline
- [ ] Docker containerization
- [ ] Kubernetes deployment
- [ ] Monitoring and observability

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Contact

For questions or issues, please open an issue on GitHub.
