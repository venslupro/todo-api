# Todo API - 部署指南

## 项目概述

Todo API 是一个功能完整的待办事项管理API，采用Go语言开发，支持以下功能：

- ✅ 用户认证和授权
- ✅ 待办事项管理（创建、读取、更新、删除）
- ✅ 团队协作功能
- ✅ 媒体文件上传到AWS S3
- ✅ 实时更新（WebSocket）
- ✅ 搜索功能
- ✅ 数据库迁移管理
- ✅ Kubernetes部署配置

## 技术栈

- **后端**: Go 1.19+
- **API**: gRPC + gRPC-Gateway (RESTful API)
- **数据库**: PostgreSQL
- **缓存**: Redis
- **存储**: AWS S3
- **部署**: Docker + Kubernetes
- **认证**: JWT

## 快速开始

### 1. 环境要求

- Go 1.19+ 
- PostgreSQL 12+
- Redis 6+
- Docker 20+
- Kubernetes 1.24+
- kubectl

### 2. 配置环境变量

复制环境变量示例文件：

```bash
cp .env.example .env
```

编辑 `.env` 文件，配置你的数据库、Redis、AWS S3等连接信息。

### 3. 本地开发

#### 启动数据库和Redis

```bash
# 使用Docker启动PostgreSQL
docker run -d --name postgres \
  -e POSTGRES_USER=todo_user \
  -e POSTGRES_PASSWORD=your_password \
  -e POSTGRES_DB=todo_api \
  -p 5432:5432 postgres:13

# 使用Docker启动Redis
docker run -d --name redis -p 6379:6379 redis:6-alpine
```

#### 运行数据库迁移

```bash
# 构建应用
go build -o main cmd/server/main.go cmd/server/migrate.go

# 运行数据库迁移
./main migrate
```

#### 启动应用

```bash
# 启动服务器
./main

# 或者直接使用go run
go run cmd/server/main.go cmd/server/migrate.go
```

应用将在以下地址启动：
- gRPC服务: localhost:50051
- REST API: http://localhost:8080

### 4. 测试API

使用curl测试API：

```bash
# 健康检查
curl http://localhost:8080/health

# 用户注册
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "password123",
    "full_name": "Test User"
  }'

# 用户登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

## Kubernetes部署

### 1. 构建Docker镜像

```bash
# 构建镜像
docker build -t your-registry/todo-api:latest .

# 推送镜像（如果需要）
docker push your-registry/todo-api:latest
```

### 2. 配置Kubernetes

更新Kubernetes配置文件中的镜像地址：

```bash
# 编辑deployment.yaml，更新镜像地址
sed -i 's|your-registry/todo-api:latest|your-actual-registry/todo-api:latest|g' deployments/deployment.yaml
```

### 3. 部署到Kubernetes

使用提供的部署脚本：

```bash
./deployment.sh
```

或者手动部署：

```bash
# 创建命名空间
kubectl apply -f deployments/namespace.yaml

# 创建配置
kubectl apply -f deployments/configmap.yaml

# 创建密钥（需要先更新secrets.yaml）
kubectl apply -f deployments/secrets.yaml

# 初始化数据库
kubectl apply -f deployments/db-init-job.yaml

# 等待数据库初始化完成
kubectl wait --for=condition=complete job/todo-api-db-init -n todo-api --timeout=300s

# 部署应用
kubectl apply -f deployments/deployment.yaml
kubectl apply -f deployments/service.yaml
```

### 4. 验证部署

```bash
# 检查Pod状态
kubectl get pods -n todo-api

# 检查服务
kubectl get services -n todo-api

# 查看日志
kubectl logs -f deployment/todo-api -n todo-api

# 端口转发访问API
kubectl port-forward -n todo-api svc/todo-api-service 8080:8080
```

## 数据库迁移

### 迁移命令

应用支持以下数据库迁移命令：

```bash
# 运行数据库迁移
./main migrate

# 检查迁移状态
./main check-migrations
```

### 迁移版本

当前支持的迁移版本：

- **001**: 创建基础表结构（用户、待办事项、团队等）
- **002**: 创建媒体附件表

## API文档

### REST API端点

- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/refresh` - 刷新令牌
- `GET /api/v1/todos` - 获取待办事项列表
- `POST /api/v1/todos` - 创建待办事项
- `GET /api/v1/todos/{id}` - 获取待办事项详情
- `PUT /api/v1/todos/{id}` - 更新待办事项
- `DELETE /api/v1/todos/{id}` - 删除待办事项
- `POST /api/v1/todos/{id}/media` - 上传媒体文件
- `GET /api/v1/teams` - 获取团队列表
- `POST /api/v1/teams` - 创建团队

### gRPC服务

gRPC服务定义在 `api/gen/todo/v1/todo.proto` 文件中。

## 测试

运行测试：

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/app/service/...

# 运行测试并显示覆盖率
go test -cover ./...
```

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查PostgreSQL是否运行
   - 验证环境变量配置
   - 检查网络连接

2. **迁移失败**
   - 确保数据库用户有创建表的权限
   - 检查迁移表是否已存在
   - 查看详细错误日志

3. **Kubernetes部署失败**
   - 检查镜像地址是否正确
   - 验证密钥配置
   - 查看Pod事件和日志

### 日志查看

```bash
# 本地开发查看日志
./main 2>&1 | tee app.log

# Kubernetes查看日志
kubectl logs -f deployment/todo-api -n todo-api

# 查看数据库初始化日志
kubectl logs job/todo-api-db-init -n todo-api
```

## 开发指南

### 项目结构

```
├── api/                    # API定义（Protocol Buffers）
├── cmd/server/             # 应用入口点
├── internal/
│   ├── app/               # 应用层
│   │   ├── handlers/      # HTTP/gRPC处理器
│   │   ├── service/       # 业务逻辑
│   │   └── routes/        # 路由定义
│   ├── domain/            # 领域模型
│   ├── infrastructure/    # 基础设施
│   └── pkg/               # 共享包
├── deployments/           # Kubernetes部署文件
└── Dockerfile            # Docker构建文件
```

### 添加新功能

1. 在 `api/gen/` 中添加或更新proto定义
2. 生成gRPC代码：`make proto`
3. 实现领域模型
4. 实现服务层逻辑
5. 实现处理器
6. 添加路由
7. 编写测试

## 许可证

本项目采用MIT许可证。

## 支持

如有问题，请提交Issue或联系开发团队。