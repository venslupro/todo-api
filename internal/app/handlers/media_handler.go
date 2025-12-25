package handlers

import (
	"context"
	"fmt"
	"io"

	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
	todov1 "github.com/venslupro/todo-api/api/gen/todo/v1"
	"github.com/venslupro/todo-api/internal/app/service"
	"github.com/venslupro/todo-api/internal/domain"
	"github.com/venslupro/todo-api/internal/pkg/middleware"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MediaHandler handles gRPC requests for media operations
type MediaHandler struct {
	todov1.UnimplementedMediaServiceServer
	mediaService *service.MediaService
}

// NewMediaHandler creates a new MediaHandler
func NewMediaHandler(mediaService *service.MediaService) *MediaHandler {
	return &MediaHandler{
		mediaService: mediaService,
	}
}

// UploadMedia handles media file upload via streaming
func (h *MediaHandler) UploadMedia(stream todov1.MediaService_UploadMediaServer) error {
	ctx := stream.Context()

	// Get current user ID from context
	_, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return grpcstatus.Error(codes.Unauthenticated, "user authentication required")
	}

	var metadata *todov1.UploadMediaRequest_Metadata
	var fileData []byte

	// Receive streaming data
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return grpcstatus.Error(codes.Internal, fmt.Sprintf("failed to receive data: %v", err))
		}

		switch data := req.Data.(type) {
		case *todov1.UploadMediaRequest_Metadata_:
			metadata = data.Metadata
		case *todov1.UploadMediaRequest_Chunk:
			fileData = append(fileData, data.Chunk...)
		}
	}

	if metadata == nil {
		return grpcstatus.Error(codes.InvalidArgument, "metadata is required")
	}

	if len(fileData) == 0 {
		return grpcstatus.Error(codes.InvalidArgument, "file data is required")
	}

	// For now, we'll use a simple approach since streaming upload to S3 is complex
	// In a production system, you might want to use direct client upload with pre-signed URLs
	return grpcstatus.Error(codes.Unimplemented, "streaming upload not yet implemented, use GenerateUploadURL instead")
}

// GenerateUploadURL generates a pre-signed URL for direct client upload
func (h *MediaHandler) GenerateUploadURL(ctx context.Context, req *todov1.GenerateUploadURLRequest) (*todov1.GenerateUploadURLResponse, error) {
	if req.Filename == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "filename is required")
	}

	if req.Type == commonv1.MediaType_MEDIA_TYPE_UNSPECIFIED {
		return nil, grpcstatus.Error(codes.InvalidArgument, "media type is required")
	}

	// Get current user ID from context
	_, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, grpcstatus.Error(codes.Unauthenticated, "user authentication required")
	}

	// For now, return a simple response
	// In a real implementation, you would generate a pre-signed S3 URL
	return &todov1.GenerateUploadURLResponse{
		UploadUrl: fmt.Sprintf("https://s3.amazonaws.com/bucket/temp/%s", req.Filename),
		MediaId:   "temp-id",
	}, nil
}

// GetMedia retrieves media metadata by ID
func (h *MediaHandler) GetMedia(ctx context.Context, req *todov1.GetMediaRequest) (*todov1.GetMediaResponse, error) {
	if req.Id == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "media id is required")
	}

	media, err := h.mediaService.GetMedia(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &todov1.GetMediaResponse{
		Media: h.domainMediaToProto(media),
	}, nil
}

// ListMedia retrieves media list with filtering and pagination
func (h *MediaHandler) ListMedia(ctx context.Context, req *todov1.ListMediaRequest) (*todov1.ListMediaResponse, error) {
	var todoID string
	if req.TodoId != nil {
		todoID = *req.TodoId
	}

	if todoID == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "todo id is required")
	}

	var pageSize, page int32
	if req.Pagination != nil {
		pageSize = req.Pagination.PageSize
		page = req.Pagination.Page
	}

	mediaList, _, err := h.mediaService.ListMedia(ctx, todoID, pageSize, page)
	if err != nil {
		return nil, err
	}

	protoMediaList := make([]*todov1.MediaAttachment, len(mediaList))
	for i, media := range mediaList {
		protoMediaList[i] = h.domainMediaToProto(media)
	}

	return &todov1.ListMediaResponse{
		Media: protoMediaList,
	}, nil
}

// DeleteMedia removes media by ID
func (h *MediaHandler) DeleteMedia(ctx context.Context, req *todov1.DeleteMediaRequest) (*todov1.DeleteMediaResponse, error) {
	if req.Id == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "media id is required")
	}

	// Get current user ID from context
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, grpcstatus.Error(codes.Unauthenticated, "user authentication required")
	}

	err = h.mediaService.DeleteMedia(ctx, req.Id, userID)
	if err != nil {
		return nil, err
	}

	return &todov1.DeleteMediaResponse{}, nil
}

// domainMediaToProto converts domain Media to protobuf MediaAttachment
func (h *MediaHandler) domainMediaToProto(media *domain.Media) *todov1.MediaAttachment {
	return &todov1.MediaAttachment{
		Id:           media.ID,
		Type:         media.FileType,
		Url:          media.FileURL,
		ThumbnailUrl: media.ThumbnailURL,
		Filename:     media.FileName,
		Size:         media.FileSize,
		Duration:     media.Duration,
		CreatedAt:    timestamppb.New(media.UploadedAt),
		TodoId:       media.TODOID,
		UserId:       media.UploadedBy,
	}
}
