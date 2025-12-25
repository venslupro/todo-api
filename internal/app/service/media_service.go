package service

import (
	"context"
	"fmt"
	"mime/multipart"

	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
	"github.com/venslupro/todo-api/internal/domain"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

// MediaRepository defines the interface for media data access
type MediaRepository interface {
	CreateMedia(ctx context.Context, media *domain.Media) error
	GetMediaByID(ctx context.Context, id string) (*domain.Media, error)
	ListMediaByTODOID(ctx context.Context, todoID string, limit, offset int) ([]*domain.Media, error)
	DeleteMedia(ctx context.Context, id string) error
	CountMediaByTODOID(ctx context.Context, todoID string) (int, error)
}

// MediaService handles media upload and management operations
type MediaService struct {
	repo    MediaRepository
	storage StorageService
}

// StorageService defines the interface for file storage operations
type StorageService interface {
	UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, userID string) (string, error)
	DeleteFile(ctx context.Context, fileURL string) error
	GetFileInfo(ctx context.Context, fileURL string) (*FileInfo, error)
}

// FileInfo contains information about a stored file
type FileInfo struct {
	URL      string
	Size     int64
	MimeType string
}

// NewMediaService creates a new MediaService
func NewMediaService(repo MediaRepository, storage StorageService) *MediaService {
	return &MediaService{
		repo:    repo,
		storage: storage,
	}
}

// UploadMedia handles media file upload
func (s *MediaService) UploadMedia(ctx context.Context, file multipart.File, header *multipart.FileHeader, todoID, userID string) (*domain.Media, error) {
	if file == nil || header == nil {
		return nil, grpcstatus.Error(codes.InvalidArgument, "file and file header are required")
	}

	if userID == "" {
		return nil, grpcstatus.Error(codes.Unauthenticated, "user authentication required")
	}

	// Upload file to storage
	fileURL, err := s.storage.UploadFile(ctx, file, header, userID)
	if err != nil {
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to upload file: %v", err))
	}

	// Determine media type based on MIME type
	mediaType := s.determineMediaType(header.Header.Get("Content-Type"))

	// Create media record
	media := domain.NewMedia(
		todoID,
		header.Filename,
		fileURL,
		mediaType,
		header.Size,
		userID,
	)

	// Set MIME type
	media.MimeType = header.Header.Get("Content-Type")

	// Save to database
	if err := s.repo.CreateMedia(ctx, media); err != nil {
		// Clean up uploaded file if database operation fails
		if delErr := s.storage.DeleteFile(ctx, fileURL); delErr != nil {
			// Log the deletion error but don't return it
			fmt.Printf("Failed to clean up uploaded file: %v", delErr)
		}
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to save media record: %v", err))
	}

	return media, nil
}

// GetMedia retrieves media by ID
func (s *MediaService) GetMedia(ctx context.Context, id string) (*domain.Media, error) {
	if id == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "media id is required")
	}

	media, err := s.repo.GetMediaByID(ctx, id)
	if err != nil {
		return nil, grpcstatus.Error(codes.NotFound, fmt.Sprintf("media not found: %v", err))
	}

	return media, nil
}

// ListMedia retrieves media list with pagination
func (s *MediaService) ListMedia(ctx context.Context, todoID string, pageSize, pageToken int32) ([]*domain.Media, int, error) {
	if todoID == "" {
		return nil, 0, grpcstatus.Error(codes.InvalidArgument, "todo id is required")
	}

	limit := int(pageSize)
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	offset := int(pageToken) * limit

	mediaList, err := s.repo.ListMediaByTODOID(ctx, todoID, limit, offset)
	if err != nil {
		return nil, 0, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to list media: %v", err))
	}

	total, err := s.repo.CountMediaByTODOID(ctx, todoID)
	if err != nil {
		return nil, 0, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to count media: %v", err))
	}

	return mediaList, total, nil
}

// DeleteMedia removes media by ID
func (s *MediaService) DeleteMedia(ctx context.Context, id, userID string) error {
	if id == "" {
		return grpcstatus.Error(codes.InvalidArgument, "media id is required")
	}

	if userID == "" {
		return grpcstatus.Error(codes.Unauthenticated, "user authentication required")
	}

	// Get media to check ownership and get file URL
	media, err := s.repo.GetMediaByID(ctx, id)
	if err != nil {
		return grpcstatus.Error(codes.NotFound, fmt.Sprintf("media not found: %v", err))
	}

	// Check if user owns the media
	if media.UploadedBy != userID {
		return grpcstatus.Error(codes.PermissionDenied, "user does not have permission to delete this media")
	}

	// Delete from storage
	if err := s.storage.DeleteFile(ctx, media.FileURL); err != nil {
		return grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to delete file from storage: %v", err))
	}

	// Delete from database
	if err := s.repo.DeleteMedia(ctx, id); err != nil {
		return grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to delete media record: %v", err))
	}

	return nil
}

// determineMediaType determines the media type based on MIME type
func (s *MediaService) determineMediaType(mimeType string) commonv1.MediaType {
	switch mimeType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return commonv1.MediaType_MEDIA_TYPE_IMAGE
	case "video/mp4", "video/quicktime", "video/x-msvideo":
		return commonv1.MediaType_MEDIA_TYPE_VIDEO
	default:
		return commonv1.MediaType_MEDIA_TYPE_UNSPECIFIED
	}
}
