package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
	"github.com/venslupro/todo-api/internal/config"
	"github.com/venslupro/todo-api/internal/domain"
)

// PostgresRepository implements TODORepository using PostgreSQL
type PostgresRepository struct {
	db *sql.DB
}

// DB returns the underlying database connection
func (r *PostgresRepository) DB() *sql.DB {
	return r.db
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(cfg *config.DatabaseConfig) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

// Close closes the database connection
func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

// Create creates a new TODO
func (r *PostgresRepository) Create(ctx context.Context, todo *domain.TODO) error {
	query := `
		INSERT INTO todos (
			id, user_id, title, description, status, priority, due_date,
			tags, is_shared, shared_by, created_at, updated_at, completed_at, assigned_to, parent_id, position
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	var dueDate, completedAt interface{}
	if todo.DueDate != nil {
		dueDate = todo.DueDate
	}
	if todo.CompletedAt != nil {
		completedAt = todo.CompletedAt
	}

	var assignedTo, parentID, sharedBy interface{}
	if todo.AssignedTo != nil {
		assignedTo = *todo.AssignedTo
	}
	if todo.ParentID != nil {
		parentID = *todo.ParentID
	}
	if todo.SharedBy != nil {
		sharedBy = *todo.SharedBy
	}

	_, err := r.db.ExecContext(ctx, query,
		todo.ID,
		todo.UserID,
		todo.Title,
		todo.Description,
		int32(todo.Status),
		int32(todo.Priority),
		dueDate,
		pq.Array(todo.Tags),
		todo.IsShared,
		sharedBy,
		todo.CreatedAt,
		todo.UpdatedAt,
		completedAt,
		assignedTo,
		parentID,
		todo.Position,
	)

	return err
}

// GetByID retrieves a TODO by ID
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*domain.TODO, error) {
	query := `
		SELECT id, user_id, title, description, status, priority, due_date,
		       tags, is_shared, shared_by, created_at, updated_at, completed_at, assigned_to, parent_id, position
		FROM todos
		WHERE id = $1
	`

	var todo domain.TODO
	var dueDate, completedAt sql.NullTime
	var assignedToStr, parentIDStr, sharedByStr sql.NullString
	var tags pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&todo.ID,
		&todo.UserID,
		&todo.Title,
		&todo.Description,
		&todo.Status,
		&todo.Priority,
		&dueDate,
		&tags,
		&todo.IsShared,
		&sharedByStr,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&completedAt,
		&assignedToStr,
		&parentIDStr,
		&todo.Position,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("todo not found: %w", err)
	}
	if err != nil {
		return nil, err
	}

	if dueDate.Valid {
		todo.DueDate = &dueDate.Time
	}
	if completedAt.Valid {
		todo.CompletedAt = &completedAt.Time
	}
	if assignedToStr.Valid {
		todo.AssignedTo = &assignedToStr.String
	}
	if parentIDStr.Valid {
		todo.ParentID = &parentIDStr.String
	}
	if sharedByStr.Valid {
		todo.SharedBy = &sharedByStr.String
	}
	todo.Tags = []string(tags)

	return &todo, nil
}

// Update updates an existing TODO
func (r *PostgresRepository) Update(ctx context.Context, todo *domain.TODO) error {
	query := `
		UPDATE todos
		SET title = $2, description = $3, status = $4, priority = $5, due_date = $6,
		    tags = $7, is_shared = $8, shared_by = $9, updated_at = $10, completed_at = $11, 
		    assigned_to = $12, parent_id = $13, position = $14
		WHERE id = $1
	`

	var dueDate, completedAt interface{}
	if todo.DueDate != nil {
		dueDate = todo.DueDate
	}
	if todo.CompletedAt != nil {
		completedAt = todo.CompletedAt
	}

	var assignedTo, parentID, sharedBy interface{}
	if todo.AssignedTo != nil {
		assignedTo = *todo.AssignedTo
	}
	if todo.ParentID != nil {
		parentID = *todo.ParentID
	}
	if todo.SharedBy != nil {
		sharedBy = *todo.SharedBy
	}

	result, err := r.db.ExecContext(ctx, query,
		todo.ID,
		todo.Title,
		todo.Description,
		int32(todo.Status),
		int32(todo.Priority),
		dueDate,
		pq.Array(todo.Tags),
		todo.IsShared,
		sharedBy,
		todo.UpdatedAt,
		completedAt,
		assignedTo,
		parentID,
		todo.Position,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("todo not found")
	}

	return nil
}

// Delete deletes a TODO by ID
func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM todos WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("todo not found")
	}

	return nil
}

// List retrieves TODOs with filtering, sorting, and pagination
func (r *PostgresRepository) List(ctx context.Context, options domain.TODOListOptions) ([]*domain.TODO, *domain.PaginationResult, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Build WHERE clause

	if len(options.Filter.IDs) > 0 {
		placeholders := make([]string, len(options.Filter.IDs))
		for i := range options.Filter.IDs {
			placeholders[i] = fmt.Sprintf("$%d", argIndex+i)
		}
		// Convert []string to []interface{}
		idArgs := make([]interface{}, len(options.Filter.IDs))
		for i, id := range options.Filter.IDs {
			idArgs[i] = id
		}
		args = append(args, idArgs...)
		conditions = append(conditions, "id IN ("+strings.Join(placeholders, ",")+")")
		argIndex += len(options.Filter.IDs)
	}

	if options.Filter.UserID != nil {
		conditions = append(conditions, "user_id = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *options.Filter.UserID)
		argIndex++
	}

	if len(options.Filter.Statuses) > 0 {
		placeholders := make([]string, len(options.Filter.Statuses))
		for i := range options.Filter.Statuses {
			placeholders[i] = fmt.Sprintf("$%d", argIndex+i)
		}
		// Convert []commonv1.Status to []interface{}
		statusArgs := make([]interface{}, len(options.Filter.Statuses))
		for i, status := range options.Filter.Statuses {
			statusArgs[i] = status
		}
		args = append(args, statusArgs...)
		conditions = append(conditions, "status IN ("+strings.Join(placeholders, ",")+")")
		argIndex += len(options.Filter.Statuses)
	}

	if len(options.Filter.Priorities) > 0 {
		placeholders := make([]string, len(options.Filter.Priorities))
		for i := range options.Filter.Priorities {
			placeholders[i] = fmt.Sprintf("$%d", argIndex+i)
		}
		// Convert []commonv1.Priority to []interface{}
		priorityArgs := make([]interface{}, len(options.Filter.Priorities))
		for i, priority := range options.Filter.Priorities {
			priorityArgs[i] = priority
		}
		args = append(args, priorityArgs...)
		conditions = append(conditions, "priority IN ("+strings.Join(placeholders, ",")+")")
		argIndex += len(options.Filter.Priorities)
	}

	if options.Filter.DueDateFrom != nil {
		conditions = append(conditions, "due_date >= $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *options.Filter.DueDateFrom)
		argIndex++
	}

	if options.Filter.DueDateTo != nil {
		conditions = append(conditions, "due_date <= $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *options.Filter.DueDateTo)
		argIndex++
	}

	if len(options.Filter.Tags) > 0 {
		conditions = append(conditions, "tags && $"+fmt.Sprintf("%d", argIndex))
		args = append(args, pq.Array(options.Filter.Tags))
		argIndex++
	}

	if options.Filter.AssignedTo != nil {
		conditions = append(conditions, "assigned_to = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *options.Filter.AssignedTo)
		argIndex++
	}

	if options.Filter.ParentID != nil {
		conditions = append(conditions, "parent_id = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *options.Filter.ParentID)
		argIndex++
	}

	if options.Filter.TeamID != nil {
		conditions = append(conditions, "team_id = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *options.Filter.TeamID)
		argIndex++
	}

	if options.Filter.IsShared != nil {
		conditions = append(conditions, "is_shared = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *options.Filter.IsShared)
		argIndex++
	}

	if options.Filter.CreatedDateFrom != nil {
		conditions = append(conditions, "created_at >= $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *options.Filter.CreatedDateFrom)
		argIndex++
	}

	if options.Filter.CreatedDateTo != nil {
		conditions = append(conditions, "created_at <= $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *options.Filter.CreatedDateTo)
		argIndex++
	}

	if options.Filter.CompletedDateFrom != nil {
		conditions = append(conditions, "completed_at >= $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *options.Filter.CompletedDateFrom)
		argIndex++
	}

	if options.Filter.CompletedDateTo != nil {
		conditions = append(conditions, "completed_at <= $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *options.Filter.CompletedDateTo)
		argIndex++
	}

	if options.Filter.SearchQuery != nil {
		searchQuery := "%" + *options.Filter.SearchQuery + "%"
		searchFields := options.Filter.SearchFields
		if len(searchFields) == 0 {
			// Default search fields
			searchFields = []string{"title", "description", "tags"}
		}

		var searchConditions []string
		for _, field := range searchFields {
			switch field {
			case "title":
				searchConditions = append(searchConditions, "title ILIKE $"+fmt.Sprintf("%d", argIndex))
			case "description":
				searchConditions = append(searchConditions, "description ILIKE $"+fmt.Sprintf("%d", argIndex))
			case "tags":
				searchConditions = append(searchConditions, "$"+fmt.Sprintf("%d", argIndex)+" = ANY(tags)")
			}
		}

		if len(searchConditions) > 0 {
			conditions = append(conditions, "("+strings.Join(searchConditions, " OR ")+")")
			args = append(args, searchQuery)
			argIndex++
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY created_at DESC"
	if len(options.SortOptions) > 0 {
		var orderParts []string
		for _, sort := range options.SortOptions {
			field := mapSortField(sort.Field)
			direction := "ASC"
			if sort.Descending {
				direction = "DESC"
			}
			orderParts = append(orderParts, fmt.Sprintf("%s %s", field, direction))
		}
		orderBy = "ORDER BY " + strings.Join(orderParts, ", ")
	}

	// Build pagination
	page := options.Page
	if page < 1 {
		page = 1
	}
	pageSize := options.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	// Count total items
	countQuery := "SELECT COUNT(*) FROM todos " + whereClause
	var totalItems int32
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalItems)
	if err != nil {
		return nil, nil, err
	}

	// Fetch items using safe string building
	var queryBuilder strings.Builder
	queryBuilder.WriteString("SELECT id, user_id, title, description, status, priority, due_date, ")
	queryBuilder.WriteString("tags, is_shared, shared_by, created_at, updated_at, completed_at, assigned_to, parent_id, position ")
	queryBuilder.WriteString("FROM todos ")
	queryBuilder.WriteString(whereClause)
	queryBuilder.WriteString(" ")
	queryBuilder.WriteString(orderBy)
	queryBuilder.WriteString(" LIMIT $")
	queryBuilder.WriteString(strconv.Itoa(argIndex))
	queryBuilder.WriteString(" OFFSET $")
	queryBuilder.WriteString(strconv.Itoa(argIndex + 1))

	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var todos []*domain.TODO
	for rows.Next() {
		var todo domain.TODO
		var dueDate, completedAt sql.NullTime
		var assignedToStr, parentIDStr, sharedByStr sql.NullString
		var tags pq.StringArray

		err := rows.Scan(
			&todo.ID,
			&todo.UserID,
			&todo.Title,
			&todo.Description,
			&todo.Status,
			&todo.Priority,
			&dueDate,
			&tags,
			&todo.IsShared,
			&sharedByStr,
			&todo.CreatedAt,
			&todo.UpdatedAt,
			&completedAt,
			&assignedToStr,
			&parentIDStr,
			&todo.Position,
		)
		if err != nil {
			return nil, nil, err
		}

		if dueDate.Valid {
			todo.DueDate = &dueDate.Time
		}
		if completedAt.Valid {
			todo.CompletedAt = &completedAt.Time
		}
		if assignedToStr.Valid {
			todo.AssignedTo = &assignedToStr.String
		}
		if parentIDStr.Valid {
			todo.ParentID = &parentIDStr.String
		}
		if sharedByStr.Valid {
			todo.SharedBy = &sharedByStr.String
		}
		todo.Tags = []string(tags)

		todos = append(todos, &todo)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, err
	}

	// Calculate pagination
	totalPages := (totalItems + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	pagination := &domain.PaginationResult{
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
		HasNext:     page < totalPages,
		HasPrev:     page > 1,
	}

	return todos, pagination, nil
}

// BulkUpdateStatus updates status for multiple TODOs
func (r *PostgresRepository) BulkUpdateStatus(ctx context.Context, ids []string, status commonv1.Status) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+1)
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	args[len(ids)] = int32(status)

	// Build parameter placeholders for the IN clause using safe concatenation
	var inClauseBuilder strings.Builder
	inClauseBuilder.WriteString("$1")
	for i := 2; i <= len(ids); i++ {
		inClauseBuilder.WriteString(", $")
		inClauseBuilder.WriteString(strconv.Itoa(i))
	}

	// Build query using safe string building
	var queryBuilder strings.Builder
	queryBuilder.WriteString("UPDATE todos SET status = $")
	queryBuilder.WriteString(strconv.Itoa(len(ids) + 1))
	queryBuilder.WriteString(", updated_at = $")
	queryBuilder.WriteString(strconv.Itoa(len(ids) + 2))
	queryBuilder.WriteString(" WHERE id IN (")
	queryBuilder.WriteString(inClauseBuilder.String())
	queryBuilder.WriteString(")")

	// Reorder args: ids first, then status, then updated_at
	newArgs := make([]interface{}, 0, len(ids)+2)
	newArgs = append(newArgs, args...)
	newArgs = append(newArgs, int32(status))
	newArgs = append(newArgs, time.Now())

	_, err := r.db.ExecContext(ctx, queryBuilder.String(), newArgs...)
	return err
}

// BulkDelete deletes multiple TODOs
func (r *PostgresRepository) BulkDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	// Use individual DELETE statements in a transaction
	// This is the safest approach as it avoids any dynamic SQL construction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	for _, id := range ids {
		_, err := tx.ExecContext(ctx, "DELETE FROM todos WHERE id = $1", id)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// Exists checks if a TODO exists by ID
func (r *PostgresRepository) Exists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM todos WHERE id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	return exists, err
}

// GetSharedTODOs retrieves shared TODOs for a team
func (r *PostgresRepository) GetSharedTODOs(ctx context.Context, teamID string, options domain.TODOListOptions) ([]*domain.TODO, *domain.PaginationResult, error) {
	// Get TODO IDs shared with this team
	sharedTODOIDs, err := r.getSharedTODOIDs(ctx, teamID)
	if err != nil {
		return nil, nil, err
	}

	if len(sharedTODOIDs) == 0 {
		return []*domain.TODO{}, &domain.PaginationResult{
			TotalItems:  0,
			TotalPages:  1,
			CurrentPage: options.Page,
			PageSize:    options.PageSize,
			HasNext:     false,
			HasPrev:     false,
		}, nil
	}

	// Add shared TODO IDs to filter
	if options.Filter.IDs == nil {
		options.Filter.IDs = sharedTODOIDs
	} else {
		// Intersection of requested IDs and shared IDs
		sharedMap := make(map[string]bool)
		for _, id := range sharedTODOIDs {
			sharedMap[id] = true
		}
		var filteredIDs []string
		for _, id := range options.Filter.IDs {
			if sharedMap[id] {
				filteredIDs = append(filteredIDs, id)
			}
		}
		options.Filter.IDs = filteredIDs
	}

	return r.List(ctx, options)
}

func (r *PostgresRepository) getSharedTODOIDs(ctx context.Context, teamID string) ([]string, error) {
	query := `SELECT todo_id FROM shared_todos WHERE team_id = $1`
	rows, err := r.db.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todoIDs []string
	for rows.Next() {
		var todoID string
		if err := rows.Scan(&todoID); err != nil {
			return nil, err
		}
		todoIDs = append(todoIDs, todoID)
	}

	return todoIDs, rows.Err()
}

// GetSharedTeams retrieves teams that a TODO is shared with
func (r *PostgresRepository) GetSharedTeams(ctx context.Context, todoID string) ([]string, error) {
	query := `SELECT team_id FROM shared_todos WHERE todo_id = $1`
	rows, err := r.db.QueryContext(ctx, query, todoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teamIDs []string
	for rows.Next() {
		var teamID string
		if err := rows.Scan(&teamID); err != nil {
			return nil, err
		}
		teamIDs = append(teamIDs, teamID)
	}

	return teamIDs, rows.Err()
}

// Migrate runs database migrations
func (r *PostgresRepository) Migrate(ctx context.Context) error {
	// Create schema_migrations table if it doesn't exist
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(50) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`

	_, err := r.db.ExecContext(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of applied migrations
	appliedMigrations := make(map[string]bool)
	rows, err := r.db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan migration version: %w", err)
		}
		appliedMigrations[version] = true
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating migrations: %w", err)
	}

	// Apply migrations in order
	migrations := []struct {
		version string
		upSQL   string
	}{
		{
			version: "001",
			upSQL: `
				-- Enable UUID extension
				CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

				-- Enable pg_trgm extension for full-text search
				CREATE EXTENSION IF NOT EXISTS pg_trgm;

				-- Users table
				CREATE TABLE IF NOT EXISTS users (
				    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				    email VARCHAR(255) UNIQUE NOT NULL,
				    username VARCHAR(100) UNIQUE NOT NULL,
				    password_hash VARCHAR(255) NOT NULL,
				    full_name VARCHAR(255),
				    avatar_url TEXT,
				    is_active BOOLEAN DEFAULT TRUE,
				    last_login_at TIMESTAMP WITH TIME ZONE,
				    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
				);

				-- TODO table (updated structure)
				CREATE TABLE IF NOT EXISTS todos (
				    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				    title VARCHAR(500) NOT NULL,
				    description TEXT,
				    status INTEGER NOT NULL DEFAULT 1, -- Maps to common.v1.Status enum
				    priority INTEGER NOT NULL DEFAULT 2, -- Maps to common.v1.Priority enum
				    due_date TIMESTAMP WITH TIME ZONE,
				    tags TEXT[] DEFAULT '{}',
				    is_shared BOOLEAN DEFAULT FALSE,
				    shared_by UUID REFERENCES users(id),
				    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				    completed_at TIMESTAMP WITH TIME ZONE,
				    assigned_to UUID REFERENCES users(id),
				    parent_id UUID REFERENCES todos(id) ON DELETE CASCADE,
				    position INTEGER NOT NULL DEFAULT 0
				);

				-- Teams table
				CREATE TABLE IF NOT EXISTS teams (
				    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				    name VARCHAR(255) NOT NULL,
				    description TEXT,
				    created_by UUID NOT NULL REFERENCES users(id),
				    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
				);

				-- Team members table
				CREATE TABLE IF NOT EXISTS team_members (
				    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
				    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				    role INTEGER NOT NULL DEFAULT 1, -- Maps to common.v1.Role enum
				    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				    PRIMARY KEY (team_id, user_id)
				);

				-- Shared TODOs table
				CREATE TABLE IF NOT EXISTS shared_todos (
				    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				    todo_id UUID NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
				    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
				    shared_by UUID NOT NULL REFERENCES users(id),
				    shared_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				    UNIQUE(todo_id, team_id)
				);

				-- Media files table
				CREATE TABLE IF NOT EXISTS media_attachments (
				    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				    todo_id UUID NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
				    file_name VARCHAR(500) NOT NULL,
				    file_url TEXT NOT NULL,
				    file_type INTEGER NOT NULL, -- Maps to common.v1.MediaType enum
				    file_size BIGINT NOT NULL,
				    mime_type VARCHAR(100),
				    thumbnail_url TEXT,
				    duration INTEGER, -- For videos
				    uploaded_by UUID NOT NULL REFERENCES users(id),
				    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
				);

				-- Activity logs table
				CREATE TABLE IF NOT EXISTS activity_logs (
				    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				    team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
				    user_id UUID NOT NULL REFERENCES users(id),
				    action VARCHAR(50) NOT NULL,
				    resource_type VARCHAR(50) NOT NULL,
				    resource_id UUID,
				    details JSONB,
				    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
				);

				-- Indexes for users table
				CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
				CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

				-- Indexes for todos table
				CREATE INDEX IF NOT EXISTS idx_todos_user_id ON todos(user_id);
				CREATE INDEX IF NOT EXISTS idx_todos_status ON todos(status);
				CREATE INDEX IF NOT EXISTS idx_todos_priority ON todos(priority);
				CREATE INDEX IF NOT EXISTS idx_todos_due_date ON todos(due_date);
				CREATE INDEX IF NOT EXISTS idx_todos_created_at ON todos(created_at);
				CREATE INDEX IF NOT EXISTS idx_todos_assigned_to ON todos(assigned_to);
				CREATE INDEX IF NOT EXISTS idx_todos_parent_id ON todos(parent_id);
				CREATE INDEX IF NOT EXISTS idx_todos_tags ON todos USING GIN(tags);
				CREATE INDEX IF NOT EXISTS idx_todos_title_trgm ON todos USING gin(title gin_trgm_ops);
				CREATE INDEX IF NOT EXISTS idx_todos_description_trgm ON todos USING gin(description gin_trgm_ops);

				-- Indexes for teams table
				CREATE INDEX IF NOT EXISTS idx_teams_created_by ON teams(created_by);

				-- Indexes for team_members table
				CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members(team_id);
				CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members(user_id);

				-- Indexes for shared_todos table
				CREATE INDEX IF NOT EXISTS idx_shared_todos_todo_id ON shared_todos(todo_id);
				CREATE INDEX IF NOT EXISTS idx_shared_todos_team_id ON shared_todos(team_id);

				-- Indexes for media_attachments table
				CREATE INDEX IF NOT EXISTS idx_media_todo_id ON media_attachments(todo_id);
				CREATE INDEX IF NOT EXISTS idx_media_uploaded_by ON media_attachments(uploaded_by);

				-- Indexes for activity_logs table
				CREATE INDEX IF NOT EXISTS idx_activity_logs_team_id ON activity_logs(team_id);
				CREATE INDEX IF NOT EXISTS idx_activity_logs_user_id ON activity_logs(user_id);
				CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at);
				CREATE INDEX IF NOT EXISTS idx_activity_logs_resource ON activity_logs(resource_type, resource_id);
			`,
		},
		{
			version: "002",
			upSQL: `
				-- Add additional indexes and constraints for media_attachments table
				CREATE INDEX IF NOT EXISTS idx_media_uploaded_at ON media_attachments(uploaded_at);
				CREATE INDEX IF NOT EXISTS idx_media_file_type ON media_attachments(file_type);
				CREATE INDEX IF NOT EXISTS idx_media_created_at ON media_attachments(created_at);
			`,
		},
	}

	// Apply migrations that haven't been applied yet
	for _, migration := range migrations {
		if appliedMigrations[migration.version] {
			continue
		}

		// Execute migration in a transaction
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", migration.version, err)
		}

		// Execute the migration SQL
		_, err = tx.ExecContext(ctx, migration.upSQL)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", migration.version, err)
		}

		// Record the migration
		_, err = tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", migration.version)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", migration.version, err)
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration.version, err)
		}

		fmt.Printf("Applied migration %s\n", migration.version)
	}

	return nil
}

// CheckMigrationStatus checks the status of database migrations
func (r *PostgresRepository) CheckMigrationStatus(ctx context.Context) error {
	// Check if schema_migrations table exists
	var tableExists bool
	err := r.db.QueryRowContext(ctx,
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'schema_migrations')").Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check migrations table: %w", err)
	}

	if !tableExists {
		return fmt.Errorf("migrations table does not exist")
	}

	// Get applied migrations
	rows, err := r.db.QueryContext(ctx, "SELECT version, applied_at FROM schema_migrations ORDER BY version")
	if err != nil {
		return fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	var appliedMigrations []string
	for rows.Next() {
		var version string
		var appliedAt time.Time
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return fmt.Errorf("failed to scan migration: %w", err)
		}
		appliedMigrations = append(appliedMigrations, fmt.Sprintf("%s (applied at %s)", version, appliedAt.Format("2006-01-02 15:04:05")))
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating migrations: %w", err)
	}

	// Expected migrations
	expectedMigrations := []string{"001", "002"}

	// Check if all expected migrations are applied
	appliedMap := make(map[string]bool)
	for _, migration := range appliedMigrations {
		// Extract version number from the formatted string
		version := strings.Split(migration, " ")[0]
		appliedMap[version] = true
	}

	var missingMigrations []string
	for _, expected := range expectedMigrations {
		if !appliedMap[expected] {
			missingMigrations = append(missingMigrations, expected)
		}
	}

	if len(missingMigrations) > 0 {
		return fmt.Errorf("missing migrations: %s", strings.Join(missingMigrations, ", "))
	}

	fmt.Printf("Migration status: OK (%d migrations applied)\n", len(appliedMigrations))
	for _, migration := range appliedMigrations {
		fmt.Printf("  - %s\n", migration)
	}

	return nil
}

// Helper functions

func mapSortField(field string) string {
	switch field {
	case "due_date", "dueDate":
		return "due_date"
	case "status":
		return "status"
	case "priority":
		return "priority"
	case "title":
		return "title"
	case "created_at", "createdAt":
		return "created_at"
	case "updated_at", "updatedAt":
		return "updated_at"
	case "completed_at", "completedAt":
		return "completed_at"
	case "assigned_to", "assignedTo":
		return "assigned_to"
	case "parent_id", "parentId":
		return "parent_id"
	case "position":
		return "position"
	default:
		return "created_at"
	}
}
