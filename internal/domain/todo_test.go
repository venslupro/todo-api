package domain

import (
	"testing"
	"time"

	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
)

func TestNewTODO(t *testing.T) {
	userID := "user-123"
	title := "Test TODO"

	todo := NewTODO(userID, title)

	if todo.ID == "" {
		t.Error("Expected TODO to have an ID")
	}
	if todo.UserID != userID {
		t.Errorf("Expected UserID to be %s, got %s", userID, todo.UserID)
	}
	if todo.Title != title {
		t.Errorf("Expected Title to be %s, got %s", title, todo.Title)
	}
	if todo.Status != commonv1.Status_STATUS_NOT_STARTED {
		t.Errorf("Expected Status to be STATUS_NOT_STARTED, got %v", todo.Status)
	}
	if todo.Priority != commonv1.Priority_PRIORITY_MEDIUM {
		t.Errorf("Expected Priority to be PRIORITY_MEDIUM, got %v", todo.Priority)
	}
	if todo.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if todo.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestTODO_IsCompleted(t *testing.T) {
	todo := NewTODO("user-123", "Test TODO")

	if todo.IsCompleted() {
		t.Error("New TODO should not be completed")
	}

	todo.Status = commonv1.Status_STATUS_COMPLETED
	if !todo.IsCompleted() {
		t.Error("TODO with STATUS_COMPLETED should be completed")
	}
}

func TestTODO_Complete(t *testing.T) {
	todo := NewTODO("user-123", "Test TODO")

	todo.Complete()

	if todo.Status != commonv1.Status_STATUS_COMPLETED {
		t.Errorf("Expected Status to be STATUS_COMPLETED, got %v", todo.Status)
	}
	if todo.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
	if todo.UpdatedAt.Before(todo.CreatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestTODO_Reopen(t *testing.T) {
	todo := NewTODO("user-123", "Test TODO")
	todo.Complete()

	todo.Reopen()

	if todo.Status != commonv1.Status_STATUS_NOT_STARTED {
		t.Errorf("Expected Status to be STATUS_NOT_STARTED, got %v", todo.Status)
	}
	if todo.CompletedAt != nil {
		t.Error("Expected CompletedAt to be cleared")
	}
}

func TestTODO_Update(t *testing.T) {
	todo := NewTODO("user-123", "Original Title")

	newTitle := "Updated Title"
	newDescription := "Updated Description"
	newStatus := commonv1.Status_STATUS_IN_PROGRESS
	newPriority := commonv1.Priority_PRIORITY_HIGH
	dueDate := time.Now().Add(24 * time.Hour)
	tags := []string{"tag1", "tag2"}
	assignedTo := "user-456"
	parentID := "parent-123"
	position := int32(5)

	todo.Update(&newTitle, &newDescription, &newStatus, &newPriority, &dueDate, tags, &assignedTo, &parentID, &position)

	if todo.Title != newTitle {
		t.Errorf("Expected Title to be %s, got %s", newTitle, todo.Title)
	}
	if todo.Description != newDescription {
		t.Errorf("Expected Description to be %s, got %s", newDescription, todo.Description)
	}
	if todo.Status != newStatus {
		t.Errorf("Expected Status to be %v, got %v", newStatus, todo.Status)
	}
	if todo.Priority != newPriority {
		t.Errorf("Expected Priority to be %v, got %v", newPriority, todo.Priority)
	}
	if todo.DueDate == nil || !todo.DueDate.Equal(dueDate) {
		t.Error("Expected DueDate to be updated")
	}
	if len(todo.Tags) != len(tags) {
		t.Errorf("Expected %d tags, got %d", len(tags), len(todo.Tags))
	}
	if todo.AssignedTo == nil || *todo.AssignedTo != assignedTo {
		t.Error("Expected AssignedTo to be updated")
	}
	if todo.ParentID == nil || *todo.ParentID != parentID {
		t.Error("Expected ParentID to be updated")
	}
	if todo.Position != position {
		t.Errorf("Expected Position to be %d, got %d", position, todo.Position)
	}
}

func TestTODO_Update_StatusToCompleted(t *testing.T) {
	todo := NewTODO("user-123", "Test TODO")
	status := commonv1.Status_STATUS_COMPLETED

	todo.Update(nil, nil, &status, nil, nil, nil, nil, nil, nil)

	if todo.Status != commonv1.Status_STATUS_COMPLETED {
		t.Errorf("Expected Status to be STATUS_COMPLETED, got %v", todo.Status)
	}
	if todo.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set when status changes to COMPLETED")
	}
}

func TestTODO_Update_StatusFromCompleted(t *testing.T) {
	todo := NewTODO("user-123", "Test TODO")
	todo.Complete()

	status := commonv1.Status_STATUS_IN_PROGRESS
	todo.Update(nil, nil, &status, nil, nil, nil, nil, nil, nil)

	if todo.CompletedAt != nil {
		t.Error("Expected CompletedAt to be cleared when status changes from COMPLETED")
	}
}
