package domain

import (
	"time"

	"github.com/google/uuid"
	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
)

// TODO represents a TODO item in the domain
type TODO struct {
	ID               string
	UserID           string
	Title            string
	Description      string
	Status           commonv1.Status
	Priority         commonv1.Priority
	DueDate          *time.Time
	Tags             []string
	IsShared         bool
	SharedBy         *string
	TeamID           *string
	MediaAttachments []MediaAttachment
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CompletedAt      *time.Time
	AssignedTo       *string
	ParentID         *string
	Position         int32
}

// MediaAttachment represents media attached to a TODO
type MediaAttachment struct {
	ID           string
	Type         commonv1.MediaType
	URL          string
	ThumbnailURL string
	Filename     string
	Size         int64
	Duration     int32
	CreatedAt    time.Time
	TODOID       string
	UserID       string
}

// NewTODO creates a new TODO with generated ID
func NewTODO(userID, title string) *TODO {
	now := time.Now()
	return &TODO{
		ID:        uuid.New().String(),
		UserID:    userID,
		Title:     title,
		Status:    commonv1.Status_STATUS_NOT_STARTED,
		Priority:  commonv1.Priority_PRIORITY_MEDIUM,
		IsShared:  false,
		CreatedAt: now,
		UpdatedAt: now,
		Position:  0,
	}
}

// IsCompleted returns true if the TODO is completed
func (t *TODO) IsCompleted() bool {
	return t.Status == commonv1.Status_STATUS_COMPLETED
}

// Complete marks the TODO as completed
func (t *TODO) Complete() {
	now := time.Now()
	t.Status = commonv1.Status_STATUS_COMPLETED
	t.CompletedAt = &now
	t.UpdatedAt = now
}

// Reopen reopens a completed TODO
func (t *TODO) Reopen() {
	t.Status = commonv1.Status_STATUS_NOT_STARTED
	t.CompletedAt = nil
	t.UpdatedAt = time.Now()
}

// Update updates TODO fields
func (t *TODO) Update(title, description *string, status *commonv1.Status, priority *commonv1.Priority, dueDate *time.Time, tags []string, assignedTo, parentID *string, position *int32) {
	if title != nil {
		t.Title = *title
	}
	if description != nil {
		t.Description = *description
	}
	if status != nil {
		t.Status = *status
		if *status == commonv1.Status_STATUS_COMPLETED && t.CompletedAt == nil {
			now := time.Now()
			t.CompletedAt = &now
		} else if *status != commonv1.Status_STATUS_COMPLETED {
			t.CompletedAt = nil
		}
	}
	if priority != nil {
		t.Priority = *priority
	}
	if dueDate != nil {
		t.DueDate = dueDate
	}
	if tags != nil {
		t.Tags = tags
	}
	if assignedTo != nil {
		t.AssignedTo = assignedTo
	}
	if parentID != nil {
		t.ParentID = parentID
	}
	if position != nil {
		t.Position = *position
	}
	t.UpdatedAt = time.Now()
}

// TODOFilter represents filtering criteria for TODO queries
type TODOFilter struct {
	IDs               []string
	UserID            *string
	Statuses          []commonv1.Status
	Priorities        []commonv1.Priority
	DueDateFrom       *time.Time
	DueDateTo         *time.Time
	CreatedDateFrom   *time.Time
	CreatedDateTo     *time.Time
	CompletedDateFrom *time.Time
	CompletedDateTo   *time.Time
	Tags              []string
	AssignedTo        *string
	ParentID          *string
	TeamID            *string
	IsShared          *bool
	SearchQuery       *string
	SearchFields      []string // Fields to search in: title, description, tags
}

// SortOption represents sorting criteria
type SortOption struct {
	Field      string
	Descending bool
}

// TODOListOptions represents options for listing TODOs
type TODOListOptions struct {
	Filter      TODOFilter
	SortOptions []SortOption
	Page        int32
	PageSize    int32
}

// PaginationResult represents pagination metadata
type PaginationResult struct {
	TotalItems  int32
	TotalPages  int32
	CurrentPage int32
	PageSize    int32
	HasNext     bool
	HasPrev     bool
}
