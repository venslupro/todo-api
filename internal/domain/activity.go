package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ActivityLog represents an activity log entry
type ActivityLog struct {
	ID           string
	TeamID       *string
	UserID       string
	Action       string
	ResourceType string
	ResourceID   *string
	Details      map[string]interface{}
	CreatedAt    time.Time
}

// NewActivityLog creates a new activity log entry
func NewActivityLog(userID, action, resourceType string, teamID, resourceID *string, details map[string]interface{}) *ActivityLog {
	return &ActivityLog{
		ID:           uuid.New().String(),
		TeamID:       teamID,
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
		CreatedAt:    time.Now(),
	}
}

// ToJSONB converts details to JSONB format for database storage
func (a *ActivityLog) ToJSONB() ([]byte, error) {
	if a.Details == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(a.Details)
}
