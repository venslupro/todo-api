# TODO API Deployment and Configuration Guide

> 📖 **Back to [README.md](../README.md)** | **Previous: [Development Guide](DEVELOPMENT.md)**

This comprehensive guide covers the deployment and configuration of TODO API across different environments, from local development to production Kubernetes clusters.

## 🏠 Local Development Deployment

### Prerequisites
- Docker and Docker Compose
- Go 1.25.5+
- Git

### Quick Start with Docker Compose

```bash
# Clone the repository
git clone https://github.com/venslupro/todo-api.git
cd todo-api

# Start all services
docker-compose up -d

# Run database migrations
docker-compose exec app go run cmd/server/migrate.go

# The application will be available at:
# HTTP API: http://localhost:8080
# gRPC API: localhost:50051
```

### Manual Local Setup

1. **Start Database and Redis**
```bash
# Start PostgreSQL
docker run -d --name postgres \
  -e POSTGRES_DB=todo_api \
  -e POSTGRES_USER=todo_user \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  postgres:13

# Start Redis
docker run -d --name redis \
  -p 6379:6379 \
  redis:6-alpine
```

2. **Configure Environment**
```bash
cp .env.example .env
# Edit .env with your database and Redis settings
```

3. **Run the Application**
```bash
# Install dependencies
go mod download

# Run migrations
go run cmd/server/migrate.go

# Start the server
go run cmd/server/main.go
```

## ☁️ Kubernetes Deployment

### Prerequisites
- Kubernetes cluster (v1.25+)
- Helm (v3.8+)
- kubectl configured
- External PostgreSQL and Redis (for production)

### Helm Chart Structure

```
deployment/kubernetes/todo-api/
├── Chart.yaml              # Chart metadata
├── values.yaml             # Default configuration
├── values-production.yaml  # Production configuration
└── templates/              # Kubernetes manifests
    ├── deployment.yaml     # Application deployment
    ├── service.yaml        # Service definition
    ├── secrets.yaml        # Secrets management
    ├── ingress.yaml        # Ingress configuration
    ├── configmap.yaml      # Configuration
    ├── migration-job.yaml  # Database migration job
    └── _helpers.tpl        # Template helpers
```

### Development Deployment

1. **Install Dependencies**
```bash
# Add Bitnami repository
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
```

2. **Deploy with Embedded Databases**
```bash
# Deploy to development namespace
helm install todo-api deployment/kubernetes/todo-api/ \
  --namespace todo-api \
  --create-namespace \
  --values deployment/kubernetes/todo-api/values.yaml
```

### Production Deployment

1. **Prepare Production Values**
```yaml
# Create production-values.yaml
app:
  image:
    tag: v1.0.0
  secrets:
    DB_PASSWORD: "your-production-db-password"
    JWT_SECRET: "your-jwt-secret"
    STORAGE_S3_KEY: "your-aws-access-key"
    STORAGE_S3_SECRET: "your-aws-secret-key"
    STORAGE_S3_BUCKET: "todo-api-production"
    STORAGE_S3_REGION: "us-east-1"

# Use external databases
postgresql:
  enabled: false
redis:
  enabled: false

# Enable ingress
ingress:
  enabled: true
  hosts:
    - host: api.yourdomain.com
```

2. **Deploy to Production**
```bash
# Create namespace
kubectl create namespace todo-api

# Create secrets (manually or via sealed-secrets)
kubectl create secret generic todo-api-secrets \
  --namespace todo-api \
  --from-literal=DB_PASSWORD=your-db-password \
  --from-literal=JWT_SECRET=your-jwt-secret \
  --from-literal=STORAGE_S3_KEY=your-aws-key \
  --from-literal=STORAGE_S3_SECRET=your-aws-secret

# Deploy the application
helm install todo-api deployment/kubernetes/todo-api/ \
  --namespace todo-api \
  --values production-values.yaml
```

## 🔧 Configuration Reference

### Database Configuration

| Variable | Default | Description | Required |
|----------|---------|-------------|----------|
| `DB_HOST` | `localhost` | PostgreSQL server hostname | Yes |
| `DB_PORT` | `5432` | PostgreSQL server port | Yes |
| `DB_USER` | `todo_user` | Database username | Yes |
| `DB_PASSWORD` | - | Database password | Yes |
| `DB_NAME` | `todo_api` | Database name | Yes |
| `DB_SSLMODE` | `disable` | SSL mode (disable/require) | No |
| `DB_MAX_CONNECTIONS` | `25` | Maximum database connections | No |
| `DB_MAX_IDLE_CONNS` | `5` | Maximum idle connections | No |
| `DB_CONN_MAX_LIFETIME` | `5m` | Connection maximum lifetime | No |
| `DB_CONN_MAX_IDLE_TIME` | `10m` | Connection maximum idle time | No |

### Server Configuration

| Variable | Default | Description | Required |
|----------|---------|-------------|----------|
| `GRPC_PORT` | `50051` | gRPC server port | Yes |
| `HTTP_PORT` | `8080` | HTTP server port | Yes |
| `ENVIRONMENT` | `development` | Environment (dev/staging/prod) | Yes |
| `LOG_LEVEL` | `info` | Log level (debug/info/warn/error) | No |
| `LOG_FORMAT` | `json` | Log format (json/text) | No |

### Authentication Configuration

| Variable | Default | Description | Required |
|----------|---------|-------------|----------|
| `JWT_SECRET` | - | JWT signing secret | Yes |
| `ACCESS_TOKEN_EXPIRY` | `15m` | Access token expiration | No |
| `REFRESH_TOKEN_EXPIRY` | `168h` | Refresh token expiration | No |
| `PASSWORD_MIN_LENGTH` | `8` | Minimum password length | No |
| `PASSWORD_REQUIRE_UPPERCASE` | `true` | Require uppercase letters | No |
| `PASSWORD_REQUIRE_LOWERCASE` | `true` | Require lowercase letters | No |
| `PASSWORD_REQUIRE_NUMBER` | `true` | Require numbers | No |
| `PASSWORD_REQUIRE_SPECIAL` | `true` | Require special characters | No |

### Redis Configuration

| Variable | Default | Description | Required |
|----------|---------|-------------|----------|
| `REDIS_HOST` | `localhost` | Redis server hostname | Yes |
| `REDIS_PORT` | `6379` | Redis server port | Yes |
| `REDIS_PASSWORD` | - | Redis password | No |
| `REDIS_DB` | `0` | Redis database number | No |

### Storage Configuration

| Variable | Default | Description | Required |
|----------|---------|-------------|----------|
| `STORAGE_TYPE` | `s3` | Storage type (s3/local) | Yes |
| `STORAGE_LOCAL_PATH` | `./uploads` | Local storage path | Conditional |
| `STORAGE_S3_BUCKET` | - | S3 bucket name | Conditional |
| `STORAGE_S3_REGION` | `us-east-1` | AWS region | Conditional |
| `STORAGE_S3_KEY` | - | AWS access key | Conditional |
| `STORAGE_S3_SECRET` | - | AWS secret key | Conditional |

## 🏠 Development Configuration

### Local Development Setup

Create a `.env` file in the project root:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=todo_user
DB_PASSWORD=password
DB_NAME=todo_api
DB_SSLMODE=disable

# Server Configuration
GRPC_PORT=50051
HTTP_PORT=8080
ENVIRONMENT=development
LOG_LEVEL=debug

# Authentication Configuration
JWT_SECRET=development-jwt-secret-change-in-production
ACCESS_TOKEN_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=168h

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379

# Storage Configuration
STORAGE_TYPE=local
```

## 🚀 Production Configuration

### Production Environment Variables

```bash
# Database Configuration
DB_HOST=production-db-host
DB_PORT=5432
DB_USER=todo_user
DB_PASSWORD=secure-production-password
DB_NAME=todo_api
DB_SSLMODE=require
DB_MAX_CONNECTIONS=50

# Server Configuration
GRPC_PORT=50051
HTTP_PORT=8080
ENVIRONMENT=production
LOG_LEVEL=info
LOG_FORMAT=json

# Authentication Configuration
JWT_SECRET=very-long-secure-jwt-secret-key
ACCESS_TOKEN_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=168h

# Redis Configuration
REDIS_HOST=production-redis-host
REDIS_PORT=6379
REDIS_PASSWORD=redis-password

# Storage Configuration
STORAGE_TYPE=s3
STORAGE_S3_BUCKET=todo-api-production
STORAGE_S3_REGION=us-east-1
STORAGE_S3_KEY=aws-access-key
STORAGE_S3_SECRET=aws-secret-key
```

## 🔒 Security Configuration

### JWT Secret Requirements
- Minimum 32 characters
- Use cryptographically secure random generation
- Store securely (Kubernetes Secrets, AWS Secrets Manager, etc.)
- Rotate regularly in production

### Database Security
- Enable SSL/TLS in production
- Use strong, unique passwords
- Implement connection pooling limits
- Regular security updates

### Storage Security
- Use IAM roles when possible
- Enable bucket encryption
- Implement proper access controls
- Regular security audits

## 📊 Monitoring and Health Checks

### Health Endpoints
- `GET /health` - Application health
- `GET /health/db` - Database connectivity
- `GET /health/redis` - Redis connectivity
- `GET /health/storage` - Storage connectivity

### Metrics Endpoints
- `GET /metrics` - Prometheus metrics
- `GET /debug/pprof` - Go profiling endpoints

## 🔧 Troubleshooting

### Common Issues

1. **Database Connection Issues**
   - Check network connectivity
   - Verify credentials
   - Ensure database is running

2. **Redis Connection Issues**
   - Verify Redis server status
   - Check authentication
   - Test network connectivity

3. **Storage Configuration Issues**
   - Verify S3 bucket permissions
   - Check AWS credentials
   - Test bucket connectivity

4. **Kubernetes Deployment Issues**
   - Check pod logs: `kubectl logs <pod-name>`
   - Verify secrets: `kubectl get secrets`
   - Check service endpoints: `kubectl get endpoints`

### Log Analysis

```bash
# View application logs
docker-compose logs app

# View Kubernetes logs
kubectl logs -l app=todo-api

# Follow logs in real-time
kubectl logs -f deployment/todo-api
```

## 📈 Performance Optimization

### Database Optimization
- Enable connection pooling
- Use appropriate indexes
- Regular vacuum and analyze
- Monitor query performance

### Redis Optimization
- Configure appropriate memory limits
- Use Redis persistence
- Monitor memory usage
- Implement cache eviction policies

### Application Optimization
- Enable gzip compression
- Configure appropriate timeouts
- Use connection pooling
- Monitor resource usage

This comprehensive guide provides everything needed to deploy and configure TODO API in any environment, from local development to production Kubernetes clusters.