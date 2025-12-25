package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

// S3Config contains AWS S3 configuration
type S3Config struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Endpoint        string // Optional: for using S3-compatible services
	UsePathStyle    bool   // Optional: for S3-compatible services
}

// S3Storage implements media storage using AWS S3
type S3Storage struct {
	config     *S3Config
	s3Client   *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}

// NewS3Storage creates a new S3 storage instance
func NewS3Storage(config *S3Config) (*S3Storage, error) {
	awsConfig := &aws.Config{
		Region: aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials(
			config.AccessKeyID,
			config.SecretAccessKey,
			"",
		),
	}

	// Configure endpoint if provided
	if config.Endpoint != "" {
		awsConfig.Endpoint = aws.String(config.Endpoint)
		awsConfig.S3ForcePathStyle = aws.Bool(config.UsePathStyle)
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %v", err)
	}

	s3Client := s3.New(sess)
	uploader := s3manager.NewUploader(sess)
	downloader := s3manager.NewDownloader(sess)

	return &S3Storage{
		config:     config,
		s3Client:   s3Client,
		uploader:   uploader,
		downloader: downloader,
	}, nil
}

// UploadFile uploads a file to S3 and returns the file URL
func (s *S3Storage) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, userID string) (string, error) {
	if file == nil {
		return "", grpcstatus.Error(codes.InvalidArgument, "file is required")
	}
	if header == nil {
		return "", grpcstatus.Error(codes.InvalidArgument, "file header is required")
	}

	// Validate file size (max 10MB)
	if header.Size > 10*1024*1024 {
		return "", grpcstatus.Error(codes.InvalidArgument, "file size exceeds 10MB limit")
	}

	// Validate file type
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/gif":       true,
		"image/webp":      true,
		"video/mp4":       true,
		"video/quicktime": true,
		"video/x-msvideo": true,
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		return "", grpcstatus.Error(codes.InvalidArgument, "unsupported file type")
	}

	// Generate unique file name
	ext := filepath.Ext(header.Filename)
	fileName := fmt.Sprintf("%s_%d%s", userID, time.Now().UnixNano(), ext)
	filePath := fmt.Sprintf("media/%s", fileName)

	// Upload file to S3
	_, err := s.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket:      aws.String(s.config.BucketName),
		Key:         aws.String(filePath),
		Body:        file,
		ACL:         aws.String("public-read"),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to upload file: %v", err))
	}

	// Generate file URL
	fileURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
		s.config.BucketName, s.config.Region, filePath)

	return fileURL, nil
}

// DeleteFile deletes a file from S3
func (s *S3Storage) DeleteFile(ctx context.Context, fileURL string) error {
	if fileURL == "" {
		return grpcstatus.Error(codes.InvalidArgument, "file URL is required")
	}

	// Extract key from URL
	key := strings.TrimPrefix(fileURL, fmt.Sprintf("https://%s.s3.%s.amazonaws.com/",
		s.config.BucketName, s.config.Region))

	_, err := s.s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to delete file: %v", err))
	}

	return nil
}

// DownloadFile downloads a file from S3
func (s *S3Storage) DownloadFile(ctx context.Context, fileURL string) (io.ReadCloser, error) {
	if fileURL == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "file URL is required")
	}

	// Extract key from URL
	key := strings.TrimPrefix(fileURL, fmt.Sprintf("https://%s.s3.%s.amazonaws.com/",
		s.config.BucketName, s.config.Region))

	result, err := s.s3Client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to download file: %v", err))
	}

	return result.Body, nil
}

// GetFileInfo retrieves file metadata from S3
func (s *S3Storage) GetFileInfo(ctx context.Context, fileURL string) (*FileInfo, error) {
	if fileURL == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "file URL is required")
	}

	// Extract key from URL
	key := strings.TrimPrefix(fileURL, fmt.Sprintf("https://%s.s3.%s.amazonaws.com/",
		s.config.BucketName, s.config.Region))

	result, err := s.s3Client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to get file info: %v", err))
	}

	return &FileInfo{
		Size:         aws.Int64Value(result.ContentLength),
		ContentType:  aws.StringValue(result.ContentType),
		LastModified: aws.TimeValue(result.LastModified),
	}, nil
}

// FileInfo contains file metadata
type FileInfo struct {
	Size         int64
	ContentType  string
	LastModified time.Time
}
