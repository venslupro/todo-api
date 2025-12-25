# TODO API - Real-time Collaborative TODO System

A scalable and well-designed real-time collaborative TODO list API application built with Go, gRPC, and gRPC-Gateway. The system supports user management, team collaboration, real-time updates, and comprehensive TODO management features.

## 🚀 Features

### Core Features
- ✅ **TODOs CRUD Operations** - Create, Read, Update, Delete TODO items with rich metadata
- ✅ **Real-time Collaboration** - WebSocket-based real-time updates for team members
- ✅ **Team Management** - Create teams, invite members, manage permissions
- ✅ **Media Attachments** - Upload and manage files associated with TODOs
- ✅ **Advanced Filtering & Search** - Multi-criteria filtering and full-text search
- ✅ **Authentication & Authorization** - JWT-based auth with role-based permissions

### Technical Features
- **gRPC & REST APIs** - Dual API interface with gRPC-Gateway
- **PostgreSQL** - Reliable data storage with migrations
- **Redis** - Caching and real-time messaging
- **S3 Storage** - Scalable file storage
- **Kubernetes Ready** - Complete Helm chart for production deployment
- **Health Checks** - Comprehensive monitoring endpoints

## 📋 API Overview

### Authentication Service
- User registration and login
- JWT token generation and refresh
- Password management with security requirements

### TODO Service
- Complete CRUD operations for TODO items
- Subtask management (parent-child relationships)
- Assignment and sharing capabilities
- Status and priority management

### Team Service
- Team creation and management
- Member invitation and role assignment
- Team-based TODO sharing

### Media Service
- File upload and management
- Image processing and thumbnails
- Secure file access

### Real-time Service
- WebSocket connections for real-time updates
- Live notifications for TODO changes
- Team collaboration features

## 🛠️ Quick Start

> 📚 **For detailed setup instructions, see the [Development Guide](docs/DEVELOPMENT.md)**

### Prerequisites
- Go 1.25.5+
- PostgreSQL 13+
- Redis 6+
- Docker (optional)

### Development Setup
```bash
# Clone the repository
git clone https://github.com/venslupro/todo-api.git
cd todo-api

# Copy environment configuration
cp .env.example .env

# Install dependencies
go mod download

# Run database migrations
go run cmd/server/migrate.go

# Start the server
go run cmd/server/main.go
```

### Production Deployment
See [deployment documentation](docs/DEPLOYMENT.md) for comprehensive deployment and configuration instructions.

## 📚 Documentation

- [Architecture Overview](docs/ARCHITECTURE.md) - System design and architecture
- [Deployment & Configuration Guide](docs/DEPLOYMENT.md) - Complete deployment and configuration instructions
- [Development Guide](docs/DEVELOPMENT.md) - Development setup and guidelines

## 🔧 Configuration

> 📋 **For complete configuration reference, see the [Deployment & Configuration Guide](docs/DEPLOYMENT.md)**

Key configuration options available in `.env`:
- Database connection settings
- JWT authentication parameters
- Redis configuration
- Storage options (S3 or local)
- Server ports and environment

## 🧪 Testing

Run the complete test suite:
```bash
go test ./... -v
```

## 📊 Monitoring

Health check endpoints:
- `GET /health` - Application health status
- `GET /metrics` - Prometheus metrics (if enabled)

## 🤝 Contributing

Please read our development guidelines in the [development documentation](docs/DEVELOPMENT.md).

## 📄 License

MIT License - see LICENSE file for details.