module github.com/venslupro/todo-api

go 1.25.5

require (
    google.golang.org/grpc v1.60.0
    google.golang.org/protobuf v1.31.0
    github.com/grpc-ecosystem/grpc-gateway/v2 v2.18.0
)

// 工具依赖
require (
    github.com/bufbuild/buf v1.28.0 // indirect
    google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.3.0 // indirect
)

// 工具版本约束
tool (
    google.golang.org/protobuf/cmd/protoc-gen-go v1.31.0
    google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.3.0
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway v2.18.0
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 v2.18.0
)