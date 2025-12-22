package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type permissionRepository struct {
	db *sql.DB
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *sql.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*Permission, error) {
	query := `
		SELECT id, code, description
		FROM auth.permissions
		WHERE id = $1
	`

	perm := &Permission{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&perm.ID,
		&perm.Code,
		&perm.Description,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("permission not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return perm, nil
}

func (r *permissionRepository) GetByCode(ctx context.Context, code string) (*Permission, error) {
	query := `
		SELECT id, code, description
		FROM auth.permissions
		WHERE code = $1
	`

	perm := &Permission{}
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&perm.ID,
		&perm.Code,
		&perm.Description,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("permission not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return perm, nil
}

func (r *permissionRepository) List(ctx context.Context) ([]*Permission, error) {
	query := `
		SELECT id, code, description
		FROM auth.permissions
		ORDER BY code
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	defer rows.Close()

	perms := []*Permission{}
	for rows.Next() {
		perm := &Permission{}
		err := rows.Scan(
			&perm.ID,
			&perm.Code,
			&perm.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		perms = append(perms, perm)
	}

	return perms, nil
}

func (r *permissionRepository) GetPermissionsByRoleID(ctx context.Context, roleID uuid.UUID) ([]*Permission, error) {
	query := `
		SELECT p.id, p.code, p.description
		FROM auth.permissions p
		INNER JOIN auth.role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.code
	`

	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions by role: %w", err)
	}
	defer rows.Close()

	perms := []*Permission{}
	for rows.Next() {
		perm := &Permission{}
		err := rows.Scan(
			&perm.ID,
			&perm.Code,
			&perm.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		perms = append(perms, perm)
	}

	return perms, nil
}

func (r *permissionRepository) GetPermissionsByUserID(ctx context.Context, userID uuid.UUID) ([]*Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.code, p.description
		FROM auth.permissions p
		INNER JOIN auth.role_permissions rp ON p.id = rp.permission_id
		INNER JOIN auth.user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY p.code
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions by user: %w", err)
	}
	defer rows.Close()

	perms := []*Permission{}
	for rows.Next() {
		perm := &Permission{}
		err := rows.Scan(
			&perm.ID,
			&perm.Code,
			&perm.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		perms = append(perms, perm)
	}

	return perms, nil
}