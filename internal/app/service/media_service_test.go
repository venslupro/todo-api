package service

import (
	"context"
	"io"
	"mime/multipart"
	"strings"
	"testing"
	"time"

	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
	"github.com/venslupro/todo-api/internal/domain"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

// mockFile is a simple mock implementation of multipart.File
type mockFile struct {
	io.Reader
}

func (m *mockFile) Close() error {
	return nil
}

func (m *mockFile) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, io.EOF
}

func (m *mockFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

// MockMediaRepository is a mock implementation of MediaRepository for testing
type MockMediaRepository struct {
	media map[string]*domain.Media
}

func NewMockMediaRepository() *MockMediaRepository {
	return &MockMediaRepository{
		media: make(map[string]*domain.Media),
	}
}

func (m *MockMediaRepository) CreateMedia(ctx context.Context, media *domain.Media) error {
	m.media[media.ID] = media
	return nil
}

func (m *MockMediaRepository) GetMediaByID(ctx context.Context, id string) (*domain.Media, error) {
	media, ok := m.media[id]
	if !ok {
		return nil, &NotFoundError{ID: id}
	}
	return media, nil
}

func (m *MockMediaRepository) ListMediaByTODOID(ctx context.Context, todoID string, limit, offset int) ([]*domain.Media, error) {
	var result []*domain.Media
	for _, media := range m.media {
		if media.TODOID == todoID {
			result = append(result, media)
		}
	}

	// Simple pagination
	if offset >= len(result) {
		return []*domain.Media{}, nil
	}

	end := offset + limit
	if end > len(result) {
		end = len(result)
	}

	return result[offset:end], nil
}

func (m *MockMediaRepository) DeleteMedia(ctx context.Context, id string) error {
	if _, ok := m.media[id]; !ok {
		return &NotFoundError{ID: id}
	}
	delete(m.media, id)
	return nil
}

func (m *MockMediaRepository) CountMediaByTODOID(ctx context.Context, todoID string) (int, error) {
	count := 0
	for _, media := range m.media {
		if media.TODOID == todoID {
			count++
		}
	}
	return count, nil
}

// MockStorageService is a mock implementation of StorageService for testing
type MockStorageService struct {
	files map[string]string // fileURL -> content
}

func NewMockStorageService() *MockStorageService {
	return &MockStorageService{
		files: make(map[string]string),
	}
}

func (m *MockStorageService) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, userID string) (string, error) {
	if file == nil || header == nil {
		return "", grpcstatus.Error(codes.InvalidArgument, "file and header are required")
	}

	if header.Size > 10*1024*1024 {
		return "", grpcstatus.Error(codes.InvalidArgument, "file size exceeds limit")
	}

	fileURL := "https://s3.amazonaws.com/bucket/" + header.Filename
	m.files[fileURL] = "mock-file-content"
	return fileURL, nil
}

func (m *MockStorageService) DeleteFile(ctx context.Context, fileURL string) error {
	if _, exists := m.files[fileURL]; !exists {
		return grpcstatus.Error(codes.NotFound, "file not found")
	}
	delete(m.files, fileURL)
	return nil
}

func (m *MockStorageService) GetFileInfo(ctx context.Context, fileURL string) (*FileInfo, error) {
	if _, exists := m.files[fileURL]; !exists {
		return nil, grpcstatus.Error(codes.NotFound, "file not found")
	}

	return &FileInfo{
		URL:      fileURL,
		Size:     1024,
		MimeType: "image/jpeg",
	}, nil
}

// TestMediaService_UploadMedia tests the UploadMedia functionality
func TestMediaService_UploadMedia(t *testing.T) {
	ctx := context.Background()
	repo := NewMockMediaRepository()
	storage := NewMockStorageService()
	service := NewMediaService(repo, storage)

	tests := []struct {
		name        string
		file        multipart.File
		header      *multipart.FileHeader
		todoID      string
		userID      string
		expectError bool
		errorCode   codes.Code
	}{
		{
			name:        "successful upload",
			file:        &mockFile{strings.NewReader("test content")},
			header:      &multipart.FileHeader{Filename: "test.jpg", Size: 1024},
			todoID:      "todo-123",
			userID:      "user-123",
			expectError: false,
		},
		{
			name:        "nil file",
			file:        nil,
			header:      &multipart.FileHeader{Filename: "test.jpg", Size: 1024},
			todoID:      "todo-123",
			userID:      "user-123",
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name:        "empty user ID",
			file:        &mockFile{strings.NewReader("test content")},
			header:      &multipart.FileHeader{Filename: "test.jpg", Size: 1024},
			todoID:      "todo-123",
			userID:      "",
			expectError: true,
			errorCode:   codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			media, err := service.UploadMedia(ctx, tt.file, tt.header, tt.todoID, tt.userID)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}

				st, ok := grpcstatus.FromError(err)
				if !ok {
					t.Errorf("error is not a gRPC status error")
					return
				}

				if st.Code() != tt.errorCode {
					t.Errorf("expected error code %v, got %v", tt.errorCode, st.Code())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if media == nil {
					t.Errorf("expected media, got nil")
					return
				}

				if media.TODOID != tt.todoID {
					t.Errorf("expected todo ID %s, got %s", tt.todoID, media.TODOID)
				}

				if media.UploadedBy != tt.userID {
					t.Errorf("expected uploaded by %s, got %s", tt.userID, media.UploadedBy)
				}
			}
		})
	}
}

// TestMediaService_GetMedia tests the GetMedia functionality
func TestMediaService_GetMedia(t *testing.T) {
	ctx := context.Background()
	repo := NewMockMediaRepository()
	storage := NewMockStorageService()
	service := NewMediaService(repo, storage)

	// Setup test data
	testMedia := &domain.Media{
		ID:         "media-123",
		TODOID:     "todo-123",
		FileName:   "test.jpg",
		FileURL:    "https://s3.amazonaws.com/bucket/test.jpg",
		FileType:   commonv1.MediaType_MEDIA_TYPE_IMAGE,
		FileSize:   1024,
		MimeType:   "image/jpeg",
		UploadedBy: "user-123",
		UploadedAt: time.Now(),
	}
	repo.CreateMedia(ctx, testMedia)

	tests := []struct {
		name        string
		mediaID     string
		expectError bool
		errorCode   codes.Code
	}{
		{
			name:        "existing media",
			mediaID:     "media-123",
			expectError: false,
		},
		{
			name:        "non-existing media",
			mediaID:     "non-existing",
			expectError: true,
			errorCode:   codes.NotFound,
		},
		{
			name:        "empty media ID",
			mediaID:     "",
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			media, err := service.GetMedia(ctx, tt.mediaID)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}

				st, ok := grpcstatus.FromError(err)
				if !ok {
					t.Errorf("error is not a gRPC status error")
					return
				}

				if st.Code() != tt.errorCode {
					t.Errorf("expected error code %v, got %v", tt.errorCode, st.Code())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if media == nil {
					t.Errorf("expected media, got nil")
					return
				}

				if media.ID != tt.mediaID {
					t.Errorf("expected media ID %s, got %s", tt.mediaID, media.ID)
				}
			}
		})
	}
}

// TestMediaService_DeleteMedia tests the DeleteMedia functionality
func TestMediaService_DeleteMedia(t *testing.T) {
	ctx := context.Background()
	repo := NewMockMediaRepository()
	storage := NewMockStorageService()
	service := NewMediaService(repo, storage)

	tests := []struct {
		name        string
		setupMedia  bool
		mediaID     string
		userID      string
		expectError bool
		errorCode   codes.Code
	}{
		{
			name:        "successful delete",
			setupMedia:  true,
			mediaID:     "media-123",
			userID:      "user-123",
			expectError: false,
		},
		{
			name:        "non-existing media",
			setupMedia:  false,
			mediaID:     "non-existing",
			userID:      "user-123",
			expectError: true,
			errorCode:   codes.NotFound,
		},
		{
			name:        "empty media ID",
			setupMedia:  false,
			mediaID:     "",
			userID:      "user-123",
			expectError: true,
			errorCode:   codes.InvalidArgument,
		},
		{
			name:        "empty user ID",
			setupMedia:  true,
			mediaID:     "media-123",
			userID:      "",
			expectError: true,
			errorCode:   codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test data if needed
			if tt.setupMedia {
				testMedia := &domain.Media{
					ID:         tt.mediaID,
					TODOID:     "todo-123",
					FileName:   "test.jpg",
					FileURL:    "https://s3.amazonaws.com/bucket/test.jpg",
					FileType:   commonv1.MediaType_MEDIA_TYPE_IMAGE,
					FileSize:   1024,
					MimeType:   "image/jpeg",
					UploadedBy: tt.userID,
					UploadedAt: time.Now(),
				}
				repo.CreateMedia(ctx, testMedia)
				// Also create a mock file in storage service
				storage.files[testMedia.FileURL] = "mock-file-content"
			}

			err := service.DeleteMedia(ctx, tt.mediaID, tt.userID)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}

				st, ok := grpcstatus.FromError(err)
				if !ok {
					t.Errorf("error is not a gRPC status error")
					return
				}

				if st.Code() != tt.errorCode {
					t.Errorf("expected error code %v, got %v", tt.errorCode, st.Code())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				// Verify media was deleted
				_, err := service.GetMedia(ctx, tt.mediaID)
				if err == nil {
					t.Errorf("media should have been deleted")
				}
			}
		})
	}
}

// TestMediaService_DetermineMediaType tests the media type determination
func TestMediaService_DetermineMediaType(t *testing.T) {
	service := &MediaService{}

	tests := []struct {
		name     string
		mimeType string
		expected commonv1.MediaType
	}{
		{
			name:     "JPEG image",
			mimeType: "image/jpeg",
			expected: commonv1.MediaType_MEDIA_TYPE_IMAGE,
		},
		{
			name:     "PNG image",
			mimeType: "image/png",
			expected: commonv1.MediaType_MEDIA_TYPE_IMAGE,
		},
		{
			name:     "MP4 video",
			mimeType: "video/mp4",
			expected: commonv1.MediaType_MEDIA_TYPE_VIDEO,
		},
		{
			name:     "unknown type",
			mimeType: "application/pdf",
			expected: commonv1.MediaType_MEDIA_TYPE_UNSPECIFIED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.determineMediaType(tt.mimeType)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
