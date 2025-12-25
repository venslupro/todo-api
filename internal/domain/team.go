package domain

import (
	"time"

	"github.com/google/uuid"
	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
)

// Team represents a team in the system
type Team struct {
	ID          string
	Name        string
	Description string
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TeamMember represents a member of a team
type TeamMember struct {
	TeamID   string
	UserID   string
	Role     commonv1.Role
	JoinedAt time.Time
}

// NewTeam creates a new team
func NewTeam(name, description, createdBy string) *Team {
	now := time.Now()
	return &Team{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// HasPermission checks if a user has the required permission level
func (tm *TeamMember) HasPermission(requiredRole commonv1.Role) bool {
	// Role hierarchy: OWNER > ADMIN > MEMBER
	roleHierarchy := map[commonv1.Role]int{
		commonv1.Role_ROLE_OWNER:  3,
		commonv1.Role_ROLE_ADMIN:  2,
		commonv1.Role_ROLE_MEMBER: 1,
	}
	return roleHierarchy[tm.Role] >= roleHierarchy[requiredRole]
}
