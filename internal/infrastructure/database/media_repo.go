package database

import (
	"context"
	"database/sql"
	"fmt"

	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
	"github.com/venslupro/todo-api/internal/domain"
)

// PostgresMediaRepository implements MediaRepository using PostgreSQL
type PostgresMediaRepository struct {
	db *sql.DB
}

// NewPostgresMediaRepository creates a new PostgreSQL media repository
func NewPostgresMediaRepository(db *sql.DB) *PostgresMediaRepository {
	return &PostgresMediaRepository{db: db}
}

// CreateMedia creates a new media attachment
func (r *PostgresMediaRepository) CreateMedia(ctx context.Context, media *domain.Media) error {
	query := `
		INSERT INTO media_attachments (
			id, todo_id, file_name, file_url, file_type, file_size, 
			mime_type, thumbnail_url, duration, uploaded_by, uploaded_at,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		media.ID,
		media.TODOID,
		media.FileName,
		media.FileURL,
		int32(media.FileType),
		media.FileSize,
		media.MimeType,
		media.ThumbnailURL,
		media.Duration,
		media.UploadedBy,
		media.UploadedAt,
		media.UploadedAt, // created_at same as uploaded_at
		media.UploadedAt, // updated_at same as uploaded_at
	)

	if err != nil {
		return fmt.Errorf("failed to create media: %w", err)
	}

	return nil
}

// GetMediaByID retrieves a media by ID
func (r *PostgresMediaRepository) GetMediaByID(ctx context.Context, id string) (*domain.Media, error) {
	query := `
		SELECT 
			id, todo_id, file_name, file_url, file_type, file_size,
			mime_type, thumbnail_url, duration, uploaded_by, uploaded_at
		FROM media_attachments 
		WHERE id = $1
	`

	var media domain.Media
	var fileType int32

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&media.ID,
		&media.TODOID,
		&media.FileName,
		&media.FileURL,
		&fileType,
		&media.FileSize,
		&media.MimeType,
		&media.ThumbnailURL,
		&media.Duration,
		&media.UploadedBy,
		&media.UploadedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("media not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get media: %w", err)
	}

	media.FileType = commonv1.MediaType(fileType)

	return &media, nil
}

// ListMediaByTODOID retrieves media attachments for a TODO with pagination
func (r *PostgresMediaRepository) ListMediaByTODOID(ctx context.Context, todoID string, limit, offset int) ([]*domain.Media, error) {
	query := `
		SELECT 
			id, todo_id, file_name, file_url, file_type, file_size,
			mime_type, thumbnail_url, duration, uploaded_by, uploaded_at
		FROM media_attachments 
		WHERE todo_id = $1
		ORDER BY uploaded_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, todoID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list media: %w", err)
	}
	defer rows.Close()

	var mediaList []*domain.Media

	for rows.Next() {
		var media domain.Media
		var fileType int32

		err := rows.Scan(
			&media.ID,
			&media.TODOID,
			&media.FileName,
			&media.FileURL,
			&fileType,
			&media.FileSize,
			&media.MimeType,
			&media.ThumbnailURL,
			&media.Duration,
			&media.UploadedBy,
			&media.UploadedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan media row: %w", err)
		}

		media.FileType = commonv1.MediaType(fileType)
		mediaList = append(mediaList, &media)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating media rows: %w", err)
	}

	return mediaList, nil
}

// DeleteMedia deletes a media attachment
func (r *PostgresMediaRepository) DeleteMedia(ctx context.Context, id string) error {
	query := `DELETE FROM media_attachments WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete media: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("media not found")
	}

	return nil
}

// CountMediaByTODOID counts media attachments for a TODO
func (r *PostgresMediaRepository) CountMediaByTODOID(ctx context.Context, todoID string) (int, error) {
	query := `SELECT COUNT(*) FROM media_attachments WHERE todo_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, todoID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count media: %w", err)
	}

	return count, nil
}
