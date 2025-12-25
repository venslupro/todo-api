package main

import (
	"context"
	"fmt"
	"log"

	"github.com/venslupro/todo-api/internal/config"
	"github.com/venslupro/todo-api/internal/infrastructure/database"
)

// migrateCommand handles database migrations
func migrateCommand() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database connection
	dbRepo, err := database.NewPostgresRepository(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbRepo.Close()

	// Run migrations
	ctx := context.Background()
	if err := dbRepo.Migrate(ctx); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	fmt.Println("Database migrations completed successfully!")
}

// checkMigrationCommand checks migration status
func checkMigrationCommand() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database connection
	dbRepo, err := database.NewPostgresRepository(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbRepo.Close()

	// Check migration status
	ctx := context.Background()
	if err := dbRepo.CheckMigrationStatus(ctx); err != nil {
		log.Fatalf("Failed to check migration status: %v", err)
	}

	fmt.Println("Database migration status: OK")
}
