# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary from builder stage
COPY --from=builder /app/main .

# Copy configuration files if any
COPY --from=builder /app/internal/config ./internal/config

# Expose ports
EXPOSE 8080 8081

# Command to run the executable
CMD ["./main"]