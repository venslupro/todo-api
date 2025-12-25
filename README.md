# TODO API - Real-time Collaborative TODO System

> **注意**: 项目文档已归档到 `docs/` 目录。请查看相应的文档目录获取详细信息。

## 📚 文档目录

- **[项目概述](docs/README.md)** - 项目功能、特性和快速开始指南
- **[系统架构](docs/architecture/)** - 架构设计、技术栈和组件说明
- **[API文档](docs/api/)** - API接口、OpenAPI规范和Postman集合
- **[部署指南](docs/deployment/)** - 部署配置、Kubernetes和Docker说明
- **[开发指南](docs/development/)** - 开发环境设置、代码结构和贡献指南

## 🚀 快速开始

```bash
# 克隆项目
git clone https://github.com/venslupro/todo-api.git
cd todo-api

# 安装依赖
go mod download

# 启动服务
go run cmd/server/main.go
```

## 📦 项目结构

```
todo-api/
├── docs/                    # 项目文档
├── api/                     # API定义和协议文件
├── cmd/                     # 应用程序入口点
├── internal/                # 内部包（不对外暴露）
├── pkg/                     # 可复用的公共包
├── deployment/              # 部署配置文件
└── scripts/                 # 构建和部署脚本
```

## 🔧 开发环境

- **Go**: 1.19+
- **PostgreSQL**: 13+
- **Redis**: 6+
- **Docker**: 20+
- **Kubernetes**: 1.25+

## 📝 许可证

本项目采用 MIT 许可证。详情请查看 [LICENSE](LICENSE) 文件。

---

**更多详细信息请查看 [docs/](docs/) 目录中的相应文档。**