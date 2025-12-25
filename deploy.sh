#!/bin/bash

# Todo API Deployment Script
# This script helps deployment the Todo API application

set -e

echo "=== Todo API Deployment Script ==="

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed or not in PATH"
    exit 1
fi

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is not installed or not in PATH"
    exit 1
fi

# Build the application
echo "Building the application..."
go build -o main cmd/server/main.go cmd/server/migrate.go

# Build Docker image
echo "Building Docker image..."
docker build -t todo-api:latest .

# Apply Kubernetes configurations
echo "Applying Kubernetes configurations..."

# Create namespace
kubectl apply -f deployments/namespace.yaml

# Create configmap
kubectl apply -f deployments/configmap.yaml

# Create secrets (you need to update this with your actual secrets)
echo "Please update deployments/secrets.yaml with your actual secrets before applying"
# kubectl apply -f deployments/secrets.yaml

# Create database initialization job
kubectl apply -f deployments/db-init-job.yaml

# Wait for database initialization to complete
echo "Waiting for database initialization to complete..."
kubectl wait --for=condition=complete job/todo-api-db-init -n todo-api --timeout=300s

# Create deployment
kubectl apply -f deployments/deployment.yaml

# Create service
kubectl apply -f deployments/service.yaml

# Create ingress (if needed)
# kubectl apply -f deployments/ingress.yaml

echo "=== Deployment completed successfully! ==="
echo ""
echo "To check the status of your deployment:"
echo "  kubectl get pods -n todo-api"
echo "  kubectl get services -n todo-api"
echo ""
echo "To view logs:"
echo "  kubectl logs -f deployment/todo-api -n todo-api"
echo ""
echo "To access the API:"
echo "  kubectl port-forward -n todo-api svc/todo-api-service 8080:8080"
echo "  Then open http://localhost:8080 in your browser"