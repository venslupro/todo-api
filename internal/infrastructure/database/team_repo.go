package database

import (
	"context"
	"database/sql"
	"fmt"

	commonv1 "github.com/venslupro/todo-api/api/gen/common/v1"
	"github.com/venslupro/todo-api/internal/domain"
)

// PostgresTeamRepository implements TeamRepository using PostgreSQL
type PostgresTeamRepository struct {
	db *sql.DB
}

// NewPostgresTeamRepository creates a new PostgreSQL team repository
func NewPostgresTeamRepository(db *sql.DB) *PostgresTeamRepository {
	return &PostgresTeamRepository{db: db}
}

// Create creates a new team
func (r *PostgresTeamRepository) Create(ctx context.Context, team *domain.Team) error {
	query := `
		INSERT INTO teams (id, name, description, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		team.ID,
		team.Name,
		team.Description,
		team.CreatedBy,
		team.CreatedAt,
		team.UpdatedAt,
	)

	return err
}

// GetByID retrieves a team by ID
func (r *PostgresTeamRepository) GetByID(ctx context.Context, id string) (*domain.Team, error) {
	query := `
		SELECT id, name, description, created_by, created_at, updated_at
		FROM teams
		WHERE id = $1
	`

	var team domain.Team
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&team.ID,
		&team.Name,
		&team.Description,
		&team.CreatedBy,
		&team.CreatedAt,
		&team.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("team not found: %w", err)
	}
	if err != nil {
		return nil, err
	}

	return &team, nil
}

// Update updates an existing team
func (r *PostgresTeamRepository) Update(ctx context.Context, team *domain.Team) error {
	query := `
		UPDATE teams
		SET name = $2, description = $3, updated_at = $4
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		team.ID,
		team.Name,
		team.Description,
		team.UpdatedAt,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("team not found")
	}

	return nil
}

// Delete deletes a team by ID
func (r *PostgresTeamRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM teams WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("team not found")
	}

	return nil
}

// ListByUser retrieves teams for a user
func (r *PostgresTeamRepository) ListByUser(ctx context.Context, userID string) ([]*domain.Team, error) {
	query := `
		SELECT t.id, t.name, t.description, t.created_by, t.created_at, t.updated_at
		FROM teams t
		INNER JOIN team_members tm ON t.id = tm.team_id
		WHERE tm.user_id = $1
		ORDER BY t.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*domain.Team
	for rows.Next() {
		var team domain.Team
		err := rows.Scan(
			&team.ID,
			&team.Name,
			&team.Description,
			&team.CreatedBy,
			&team.CreatedAt,
			&team.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		teams = append(teams, &team)
	}

	return teams, rows.Err()
}

// AddMember adds a member to a team
func (r *PostgresTeamRepository) AddMember(ctx context.Context, member *domain.TeamMember) error {
	query := `
		INSERT INTO team_members (team_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (team_id, user_id) DO UPDATE SET role = $3
	`

	_, err := r.db.ExecContext(ctx, query,
		member.TeamID,
		member.UserID,
		int32(member.Role),
		member.JoinedAt,
	)

	return err
}

// RemoveMember removes a member from a team
func (r *PostgresTeamRepository) RemoveMember(ctx context.Context, teamID, userID string) error {
	query := `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`

	_, err := r.db.ExecContext(ctx, query, teamID, userID)
	return err
}

// GetMember retrieves a team member
func (r *PostgresTeamRepository) GetMember(ctx context.Context, teamID, userID string) (*domain.TeamMember, error) {
	query := `
		SELECT team_id, user_id, role, joined_at
		FROM team_members
		WHERE team_id = $1 AND user_id = $2
	`

	var member domain.TeamMember
	var roleInt int32

	err := r.db.QueryRowContext(ctx, query, teamID, userID).Scan(
		&member.TeamID,
		&member.UserID,
		&roleInt,
		&member.JoinedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("team member not found: %w", err)
	}
	if err != nil {
		return nil, err
	}

	member.Role = commonv1.Role(roleInt)

	return &member, nil
}

// ListMembers retrieves all members of a team
func (r *PostgresTeamRepository) ListMembers(ctx context.Context, teamID string) ([]*domain.TeamMember, error) {
	query := `
		SELECT team_id, user_id, role, joined_at
		FROM team_members
		WHERE team_id = $1
		ORDER BY joined_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*domain.TeamMember
	for rows.Next() {
		var member domain.TeamMember
		var roleInt int32
		err := rows.Scan(
			&member.TeamID,
			&member.UserID,
			&roleInt,
			&member.JoinedAt,
		)
		if err != nil {
			return nil, err
		}
		member.Role = commonv1.Role(roleInt)
		members = append(members, &member)
	}

	return members, rows.Err()
}

// UpdateMemberRole updates a member's role
func (r *PostgresTeamRepository) UpdateMemberRole(ctx context.Context, teamID, userID string, role commonv1.Role) error {
	query := `UPDATE team_members SET role = $3 WHERE team_id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, teamID, userID, int32(role))
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("team member not found")
	}

	return nil
}

// ShareTODO shares a TODO with a team
func (r *PostgresTeamRepository) ShareTODO(ctx context.Context, todoID, teamID, sharedBy string) error {
	query := `
		INSERT INTO shared_todos (todo_id, team_id, shared_by, shared_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (todo_id, team_id) DO UPDATE SET shared_by = $3, shared_at = NOW()
	`

	_, err := r.db.ExecContext(ctx, query, todoID, teamID, sharedBy)
	return err
}

// UnshareTODO unshares a TODO from a team
func (r *PostgresTeamRepository) UnshareTODO(ctx context.Context, todoID, teamID string) error {
	query := `DELETE FROM shared_todos WHERE todo_id = $1 AND team_id = $2`

	_, err := r.db.ExecContext(ctx, query, todoID, teamID)
	return err
}

// GetSharedTODOs retrieves TODOs shared with a team
func (r *PostgresTeamRepository) GetSharedTODOs(ctx context.Context, teamID string) ([]string, error) {
	query := `SELECT todo_id FROM shared_todos WHERE team_id = $1`

	rows, err := r.db.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todoIDs []string
	for rows.Next() {
		var todoID string
		err := rows.Scan(&todoID)
		if err != nil {
			return nil, err
		}
		todoIDs = append(todoIDs, todoID)
	}

	return todoIDs, rows.Err()
}

// GetSharedTeams retrieves teams that a TODO is shared with
func (r *PostgresTeamRepository) GetSharedTeams(ctx context.Context, todoID string) ([]string, error) {
	query := `SELECT team_id FROM shared_todos WHERE todo_id = $1`

	rows, err := r.db.QueryContext(ctx, query, todoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teamIDs []string
	for rows.Next() {
		var teamID string
		err := rows.Scan(&teamID)
		if err != nil {
			return nil, err
		}
		teamIDs = append(teamIDs, teamID)
	}

	return teamIDs, rows.Err()
}
