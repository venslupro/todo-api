package domain

import (
	"context"

	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
)

// TODORepository defines the interface for TODO data access
type TODORepository interface {
	// Create creates a new TODO
	Create(ctx context.Context, todo *TODO) error

	// GetByID retrieves a TODO by ID
	GetByID(ctx context.Context, id string) (*TODO, error)

	// Update updates an existing TODO
	Update(ctx context.Context, todo *TODO) error

	// Delete deletes a TODO by ID
	Delete(ctx context.Context, id string) error

	// List retrieves TODOs with filtering, sorting, and pagination
	List(ctx context.Context, options TODOListOptions) ([]*TODO, *PaginationResult, error)

	// BulkUpdateStatus updates status for multiple TODOs
	BulkUpdateStatus(ctx context.Context, ids []string, status commonv1.Status) error

	// BulkDelete deletes multiple TODOs
	BulkDelete(ctx context.Context, ids []string) error

	// Exists checks if a TODO exists by ID
	Exists(ctx context.Context, id string) (bool, error)

	// GetSharedTODOs retrieves shared TODOs for a team
	GetSharedTODOs(ctx context.Context, teamID string, options TODOListOptions) ([]*TODO, *PaginationResult, error)

	// GetSharedTeams retrieves teams that a TODO is shared with
	GetSharedTeams(ctx context.Context, todoID string) ([]string, error)
}

// UserRepository defines the interface for User data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id string) (*User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*User, error)

	// GetByUsername retrieves a user by username
	GetByUsername(ctx context.Context, username string) (*User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *User) error

	// Exists checks if a user exists by email
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// Exists checks if a user exists by username
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}

// TeamRepository defines the interface for Team data access
type TeamRepository interface {
	// Create creates a new team
	Create(ctx context.Context, team *Team) error

	// GetByID retrieves a team by ID
	GetByID(ctx context.Context, id string) (*Team, error)

	// Update updates an existing team
	Update(ctx context.Context, team *Team) error

	// Delete deletes a team by ID
	Delete(ctx context.Context, id string) error

	// ListByUser retrieves teams for a user
	ListByUser(ctx context.Context, userID string) ([]*Team, error)

	// AddMember adds a member to a team
	AddMember(ctx context.Context, member *TeamMember) error

	// RemoveMember removes a member from a team
	RemoveMember(ctx context.Context, teamID, userID string) error

	// GetMember retrieves a team member
	GetMember(ctx context.Context, teamID, userID string) (*TeamMember, error)

	// ListMembers retrieves all members of a team
	ListMembers(ctx context.Context, teamID string) ([]*TeamMember, error)

	// UpdateMemberRole updates a member's role
	UpdateMemberRole(ctx context.Context, teamID, userID string, role commonv1.Role) error

	// ShareTODO shares a TODO with a team
	ShareTODO(ctx context.Context, todoID, teamID, sharedBy string) error

	// UnshareTODO unshares a TODO from a team
	UnshareTODO(ctx context.Context, todoID, teamID string) error

	// GetSharedTODOs retrieves TODOs shared with a team
	GetSharedTODOs(ctx context.Context, teamID string) ([]string, error)

	// GetSharedTeams retrieves teams that a TODO is shared with
	GetSharedTeams(ctx context.Context, todoID string) ([]string, error)
}

// MediaRepository defines the interface for Media data access
type MediaRepository interface {
	// CreateMedia creates a new media attachment
	CreateMedia(ctx context.Context, media *Media) error

	// GetMediaByID retrieves a media by ID
	GetMediaByID(ctx context.Context, id string) (*Media, error)

	// ListMediaByTODOID retrieves media attachments for a TODO with pagination
	ListMediaByTODOID(ctx context.Context, todoID string, limit, offset int) ([]*Media, error)

	// DeleteMedia deletes a media attachment
	DeleteMedia(ctx context.Context, id string) error

	// CountMediaByTODOID counts media attachments for a TODO
	CountMediaByTODOID(ctx context.Context, todoID string) (int, error)
}

// ActivityRepository defines the interface for Activity log data access
type ActivityRepository interface {
	// Create creates a new activity log entry
	Create(ctx context.Context, log *ActivityLog) error

	// ListByTeam retrieves activity logs for a team
	ListByTeam(ctx context.Context, teamID string, limit int) ([]*ActivityLog, error)

	// ListByUser retrieves activity logs for a user
	ListByUser(ctx context.Context, userID string, limit int) ([]*ActivityLog, error)
}
