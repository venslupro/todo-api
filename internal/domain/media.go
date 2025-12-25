package domain

import (
	"time"

	"github.com/google/uuid"
	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
)

// Media represents a media file attached to a TODO
type Media struct {
	ID           string
	TODOID       string
	FileName     string
	FileURL      string
	FileType     commonv1.MediaType
	FileSize     int64
	MimeType     string
	ThumbnailURL string
	Duration     int32
	UploadedBy   string
	UploadedAt   time.Time
}

// NewMedia creates a new media attachment
func NewMedia(todoID, fileName, fileURL string, fileType commonv1.MediaType, fileSize int64, uploadedBy string) *Media {
	return &Media{
		ID:         uuid.New().String(),
		TODOID:     todoID,
		FileName:   fileName,
		FileURL:    fileURL,
		FileType:   fileType,
		FileSize:   fileSize,
		UploadedBy: uploadedBy,
		UploadedAt: time.Now(),
	}
}
