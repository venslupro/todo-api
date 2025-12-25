# Todo API Helm Chart

A Helm chart for deploying the Todo API application on Kubernetes.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.8+
- PostgreSQL database (can be deployed with this chart)
- Redis cache (can be deployed with this chart)
- AWS S3 bucket for media storage

## Installation

### Add the repository

```bash
helm repo add todo-api https://venslupro.github.io/todo-api
helm repo update
```

### Install the chart

```bash
# Install with default values
helm install todo-api todo-api/todo-api --namespace todo-api --create-namespace

# Install with custom values
helm install todo-api todo-api/todo-api --namespace todo-api --create-namespace \
  --set app.image.tag=v1.0.0 \
  --set app.replicaCount=3 \
  --set app.secrets.DB_PASSWORD=your-db-password \
  --set app.secrets.JWT_SECRET=your-jwt-secret
```

### Using a values file

Create a `values.yaml` file:

```yaml
app:
  replicaCount: 3
  image:
    repository: ghcr.io/venslupro/todo-api
    tag: v1.0.0
  secrets:
    DB_PASSWORD: "your-db-password"
    JWT_SECRET: "your-jwt-secret"
    STORAGE_S3_BUCKET: "your-s3-bucket"
    STORAGE_S3_REGION: "us-east-1"
    STORAGE_S3_KEY: "your-access-key"
    STORAGE_S3_SECRET: "your-secret-key"

postgresql:
  enabled: true
  auth:
    postgresPassword: "your-postgres-password"

redis:
  enabled: true
```

Then install:

```bash
helm install todo-api todo-api/todo-api --namespace todo-api --create-namespace -f values.yaml
```

## Configuration

### Application Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `app.replicaCount` | Number of replicas | `3` |
| `app.image.repository` | Docker image repository | `ghcr.io/venslupro/todo-api` |
| `app.image.tag` | Docker image tag | `latest` |
| `app.image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `app.service.type` | Service type | `ClusterIP` |
| `app.service.port` | HTTP service port | `8080` |
| `app.service.grpcPort` | gRPC service port | `50051` |
| `app.resources.requests` | Resource requests | `memory: 256Mi, cpu: 250m` |
| `app.resources.limits` | Resource limits | `memory: 512Mi, cpu: 500m` |

### Database Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `postgresql.enabled` | Enable PostgreSQL deployment | `true` |
| `postgresql.auth.postgresPassword` | PostgreSQL password | `""` |
| `postgresql.persistence.enabled` | Enable persistence | `true` |
| `postgresql.persistence.size` | Persistent volume size | `8Gi` |

### Redis Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `redis.enabled` | Enable Redis deployment | `true` |
| `redis.architecture` | Redis architecture | `replication` |
| `redis.master.persistence.enabled` | Enable persistence | `true` |
| `redis.master.persistence.size` | Persistent volume size | `1Gi` |

### Secrets Configuration

**Important**: These values should be provided via Kubernetes Secrets or external secret management.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `app.secrets.DB_PASSWORD` | Database password | `""` |
| `app.secrets.JWT_SECRET` | JWT secret key | `""` |
| `app.secrets.STORAGE_S3_BUCKET` | AWS S3 bucket name | `""` |
| `app.secrets.STORAGE_S3_REGION` | AWS S3 region | `""` |
| `app.secrets.STORAGE_S3_KEY` | AWS S3 access key | `""` |
| `app.secrets.STORAGE_S3_SECRET` | AWS S3 secret key | `""` |

### Ingress Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable ingress | `false` |
| `ingress.className` | Ingress class name | `""` |
| `ingress.hosts` | Ingress hosts configuration | `[]` |

### Autoscaling Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `autoscaling.enabled` | Enable autoscaling | `false` |
| `autoscaling.minReplicas` | Minimum replicas | `1` |
| `autoscaling.maxReplicas` | Maximum replicas | `10` |
| `autoscaling.targetCPUUtilizationPercentage` | CPU utilization target | `80` |

## Environment Variables

The application uses the following environment variables:

### Database
- `DB_HOST`: Database host
- `DB_PORT`: Database port
- `DB_NAME`: Database name
- `DB_USER`: Database user
- `DB_PASSWORD`: Database password
- `DB_SSLMODE`: SSL mode

### Redis
- `REDIS_HOST`: Redis host
- `REDIS_PORT`: Redis port

### Server
- `GRPC_PORT`: gRPC server port
- `HTTP_PORT`: HTTP server port
- `ENVIRONMENT`: Application environment

### Authentication
- `JWT_SECRET`: JWT signing secret
- `ACCESS_TOKEN_EXPIRY`: Access token expiry
- `REFRESH_TOKEN_EXPIRY`: Refresh token expiry

### Storage
- `STORAGE_TYPE`: Storage type (s3)
- `STORAGE_S3_BUCKET`: S3 bucket name
- `STORAGE_S3_REGION`: S3 region
- `STORAGE_S3_KEY`: S3 access key
- `STORAGE_S3_SECRET`: S3 secret key

## Database Migration

The chart includes a database migration job that runs before the application deployment:

```yaml
migration:
  enabled: true
  image:
    repository: ghcr.io/venslupro/todo-api
    tag: latest
```

The migration job uses Helm hooks to ensure it runs before the application starts.

## Health Checks

The application includes liveness and readiness probes:

```yaml
probes:
  liveness:
    path: /health
    initialDelaySeconds: 30
    periodSeconds: 10
  readiness:
    path: /health
    initialDelaySeconds: 5
    periodSeconds: 5
```

## Security

### Pod Security Context

```yaml
podSecurityContext:
  fsGroup: 1000
  runAsUser: 1000
  runAsNonRoot: true

securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 1000
```

### Service Account

A dedicated service account is created for the application:

```yaml
serviceAccount:
  create: true
  name: ""
```

## Monitoring

### Resource Monitoring

Resource requests and limits can be configured:

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### Autoscaling

Horizontal Pod Autoscaler can be enabled:

```yaml
autoscaling:
  enabled: true
  minReplicas: 1
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
```

## Network Configuration

### Service

The application exposes two ports:
- HTTP: 8080 (REST API)
- gRPC: 50051 (gRPC API)

```yaml
service:
  type: ClusterIP
  port: 8080
  grpcPort: 50051
```

### Ingress

Ingress can be configured for external access:

```yaml
ingress:
  enabled: true
  className: nginx
  hosts:
    - host: todo-api.example.com
      paths:
        - path: /
          pathType: Prefix
```

## Storage

### Database Persistence

PostgreSQL persistence is enabled by default:

```yaml
postgresql:
  persistence:
    enabled: true
    size: 8Gi
```

### Redis Persistence

Redis persistence is enabled by default:

```yaml
redis:
  master:
    persistence:
      enabled: true
      size: 1Gi
```

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -n todo-api
```

### Check Logs

```bash
kubectl logs -f deployment/todo-api -n todo-api
```

### Check Database Connection

```bash
kubectl exec -it deployment/todo-api -n todo-api -- /app/main migrate --check
```

### Port Forward for Local Testing

```bash
kubectl port-forward svc/todo-api 8080:8080 -n todo-api
```

Then access the API at `http://localhost:8080`

## Uninstalling

To uninstall the chart:

```bash
helm uninstall todo-api -n todo-api
kubectl delete namespace todo-api
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test the chart locally
5. Submit a pull request

## License

This chart is licensed under the MIT License.