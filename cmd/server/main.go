package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	todov1 "github.com/venslupro/todo-api/api/gen/todo/v1"
	"github.com/venslupro/todo-api/internal/app/handlers"
	"github.com/venslupro/todo-api/internal/app/routes"
	"github.com/venslupro/todo-api/internal/app/service"
	"github.com/venslupro/todo-api/internal/config"
	"github.com/venslupro/todo-api/internal/infrastructure/database"
	"github.com/venslupro/todo-api/internal/infrastructure/redis"
	"github.com/venslupro/todo-api/internal/pkg/auth"
)

func main() {
	// Check for migration commands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			migrateCommand()
			return
		case "check-migrations":
			checkMigrationCommand()
			return
		}
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	dbRepo, err := database.NewPostgresRepository(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbRepo.Close()

	// Initialize Redis
	redisClient, err := redis.NewClient(redis.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize repositories
	userRepo := database.NewPostgresUserRepository(dbRepo.DB())
	teamRepo := database.NewPostgresTeamRepository(dbRepo.DB())
	todoRepo := dbRepo
	_ = redis.NewCacheRepository(redisClient) // cacheRepo - will be used when caching is implemented

	// Initialize JWT manager
	jwtMgr := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.AccessTokenExpiry)

	// Initialize WebSocket service
	websocketService := service.NewWebSocketService()

	// Initialize services
	authService := service.NewAuthService(userRepo, jwtMgr)
	teamService := service.NewTeamService(teamRepo, websocketService)
	todoService := service.NewTODOService(todoRepo, websocketService)

	// Initialize handlers
	todoHandler := handlers.NewTODOHandler(todoService)
	websocketHandler := handlers.NewWebSocketHandler(websocketService, authService, teamService)

	// Start WebSocket service
	go websocketService.Run()

	// Start gRPC server
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port: %v", err)
	}

	grpcServer := grpc.NewServer()
	todov1.RegisterTODOServiceServer(grpcServer, todoHandler)

	// Start gRPC server in a goroutine
	go func() {
		log.Printf("Starting gRPC server on port %d", cfg.Server.GRPCPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Start HTTP server (gRPC-Gateway + WebSocket)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create main HTTP mux
	httpMux := http.NewServeMux()

	// Create gRPC-Gateway mux
	gatewayMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err = todov1.RegisterTODOServiceHandlerFromEndpoint(ctx, gatewayMux, fmt.Sprintf("localhost:%d", cfg.Server.GRPCPort), opts)
	if err != nil {
		log.Fatalf("Failed to register gateway: %v", err)
	}

	// Mount gRPC-Gateway under /v1/
	httpMux.Handle("/v1/", gatewayMux)

	// Register WebSocket routes
	routes.RegisterWebSocketRoutes(httpMux, websocketHandler)

	// Register search routes
	routes.RegisterSearchRoutes(httpMux, todoHandler)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler: httpMux,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("Starting HTTP server on port %d", cfg.Server.HTTPPort)
		log.Printf("API available at http://localhost:%d/v1/", cfg.Server.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to serve HTTP: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	// Graceful shutdown
	cancel()
	grpcServer.GracefulStop()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
	}

	log.Println("Servers stopped")
}
