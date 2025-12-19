package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type userRoleRepository struct {
	db *sql.DB
}

// NewUserRoleRepository creates a new user-role repository
func NewUserRoleRepository(db *sql.DB) UserRoleRepository {
	return &userRoleRepository{db: db}
}

func (r *userRoleRepository) Assign(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `
		INSERT INTO auth.user_roles (user_id, role_id, assigned_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, userID, roleID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

func (r *userRoleRepository) Revoke(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `
		DELETE FROM auth.user_roles
		WHERE user_id = $1 AND role_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user role assignment not found")
	}

	return nil
}

func (r *userRoleRepository) GetRolesByUserID(ctx context.Context, userID uuid.UUID) ([]*Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.created_at
		FROM auth.roles r
		INNER JOIN auth.user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles by user: %w", err)
	}
	defer rows.Close()

	roles := []*Role{}
	for rows.Next() {
		role := &Role{}
		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *userRoleRepository) GetUsersByRoleID(ctx context.Context, roleID uuid.UUID) ([]*User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.is_active, u.created_at, u.last_login
		FROM auth.users u
		INNER JOIN auth.user_roles ur ON u.id = ur.user_id
		WHERE ur.role_id = $1
		ORDER BY u.username
	`

	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by role: %w", err)
	}
	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.IsActive,
			&user.CreatedAt,
			&user.LastLogin,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}