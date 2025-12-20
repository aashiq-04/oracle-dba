package graph

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/aashiq-04/oracle-dba/internal/graph/model"
	"github.com/aashiq-04/oracle-dba/internal/middleware"
	"github.com/aashiq-04/oracle-dba/internal/repository"
)

// ============================================================================
// QUERY RESOLVERS - Authentication
// ============================================================================

// Me returns the currently authenticated user
func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
	// Require authentication
	userCtx, err := middleware.RequireAuth(ctx)
	if err != nil {
		return nil, err
	}

	// Get user from repository
	user, err := r.authService.GetUserByID(ctx, userCtx.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get user roles
	roles, err := r.rbacService.GetUserRoles(ctx, userCtx.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return mapUserToGraphQL(user, roles), nil
}

// Users returns all users (Admin only)
func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	// Require MANAGE_USERS permission
	if err := middleware.RequirePermission(ctx, "MANAGE_USERS"); err != nil {
		return nil, err
	}

	// Implementation would go here
	return nil, fmt.Errorf("not implemented")
}

// User returns a specific user by ID (Admin only)
func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
	// Require MANAGE_USERS permission
	if err := middleware.RequirePermission(ctx, "MANAGE_USERS"); err != nil {
		return nil, err
	}

	// Implementation would go here
	return nil, fmt.Errorf("not implemented")
}

// Roles returns all available roles
func (r *queryResolver) Roles(ctx context.Context) ([]*model.Role, error) {
	// Require authentication
	if _, err := middleware.RequireAuth(ctx); err != nil {
		return nil, err
	}

	roles, err := r.rbacService.GetAllRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	return mapRolesToGraphQL(roles), nil
}

// Permissions returns all available permissions
func (r *queryResolver) Permissions(ctx context.Context) ([]*model.Permission, error) {
	// Require authentication
	if _, err := middleware.RequireAuth(ctx); err != nil {
		return nil, err
	}

	permissions, err := r.rbacService.GetAllPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	return mapPermissionsToGraphQL(permissions), nil
}

// ============================================================================
// MUTATION RESOLVERS - Authentication
// ============================================================================

// Login authenticates a user and returns a JWT token
func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (*model.AuthPayload, error) {
	// Perform login
	loginResp, err := r.authService.Login(ctx, input.Username, input.Password)
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	// Map to GraphQL response
	return &model.AuthPayload{
		Token:     loginResp.Token,
		User:      mapUserToGraphQL(loginResp.User, loginResp.Roles),
		ExpiresAt: loginResp.ExpiresAt,
	}, nil
}

// Logout logs out the current user (currently just returns true)
func (r *mutationResolver) Logout(ctx context.Context) (bool, error) {
	// Require authentication
	if _, err := middleware.RequireAuth(ctx); err != nil {
		return false, err
	}

	// In a JWT-based system, logout is typically handled client-side
	// by removing the token. We just return success here.
	return true, nil
}

// CreateUser creates a new user (Admin only)
func (r *mutationResolver) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
	// Require MANAGE_USERS permission
	if err := middleware.RequirePermission(ctx, "MANAGE_USERS"); err != nil {
		return nil, err
	}

	// Parse role IDs
	roleIDs := make([]uuid.UUID, len(input.RoleIds))
	for i, idStr := range input.RoleIds {
		roleID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid role ID: %w", err)
		}
		roleIDs[i] = roleID
	}

	// Create user
	user, err := r.authService.CreateUser(ctx, input.Username, input.Email, input.Password, roleIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Get user roles
	roles, err := r.rbacService.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return mapUserToGraphQL(user, roles), nil
}

// UpdateUser updates a user (Admin only)
func (r *mutationResolver) UpdateUser(ctx context.Context, input model.UpdateUserInput) (*model.User, error) {
	// Require MANAGE_USERS permission
	if err := middleware.RequirePermission(ctx, "MANAGE_USERS"); err != nil {
		return nil, err
	}

	// Implementation would go here
	return nil, fmt.Errorf("not implemented")
}

// DeleteUser deletes a user (Admin only)
func (r *mutationResolver) DeleteUser(ctx context.Context, userID string) (bool, error) {
	// Require MANAGE_USERS permission
	if err := middleware.RequirePermission(ctx, "MANAGE_USERS"); err != nil {
		return false, err
	}

	// Implementation would go here
	return false, fmt.Errorf("not implemented")
}

// AssignRole assigns a role to a user (Admin only)
func (r *mutationResolver) AssignRole(ctx context.Context, userID string, roleID string) (*model.User, error) {
	// Require MANAGE_USERS permission
	if err := middleware.RequirePermission(ctx, "MANAGE_USERS"); err != nil {
		return nil, err
	}

	// Get current user for audit
	userCtx := middleware.MustGetUserFromContext(ctx)

	// Parse IDs
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	rid, err := uuid.Parse(roleID)
	if err != nil {
		return nil, fmt.Errorf("invalid role ID: %w", err)
	}

	// Assign role
	if err := r.rbacService.AssignRole(ctx, uid, rid, userCtx.UserID); err != nil {
		return nil, fmt.Errorf("failed to assign role: %w", err)
	}

	// Implementation would return updated user
	return nil, fmt.Errorf("not implemented - return updated user")
}

// RevokeRole revokes a role from a user (Admin only)
func (r *mutationResolver) RevokeRole(ctx context.Context, userID string, roleID string) (*model.User, error) {
	// Require MANAGE_USERS permission
	if err := middleware.RequirePermission(ctx, "MANAGE_USERS"); err != nil {
		return nil, err
	}

	// Get current user for audit
	userCtx := middleware.MustGetUserFromContext(ctx)

	// Parse IDs
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	rid, err := uuid.Parse(roleID)
	if err != nil {
		return nil, fmt.Errorf("invalid role ID: %w", err)
	}

	// Revoke role
	if err := r.rbacService.RevokeRole(ctx, uid, rid, userCtx.UserID); err != nil {
		return nil, fmt.Errorf("failed to revoke role: %w", err)
	}

	// Implementation would return updated user
	return nil, fmt.Errorf("not implemented - return updated user")
}

// ============================================================================
// HELPER FUNCTIONS - Mappers
// ============================================================================

// mapUserToGraphQL converts repository user to GraphQL model
func mapUserToGraphQL(user *repository.User, roles []*repository.Role) *model.User {
	graphqlRoles := make([]*model.Role, len(roles))
	for i, role := range roles {
		graphqlRoles[i] = &model.Role{
			ID:          role.ID.String(),
			Name:        role.Name,
			Description: role.Description,
		}
	}

	return &model.User{
		ID:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		IsActive:  user.IsActive,
		Roles:     graphqlRoles,
		LastLogin: user.LastLogin,
		CreatedAt: user.CreatedAt,
	}
}

// mapRolesToGraphQL converts repository roles to GraphQL models
func mapRolesToGraphQL(roles []*repository.Role) []*model.Role {
	result := make([]*model.Role, len(roles))
	for i, role := range roles {
		result[i] = &model.Role{
			ID:          role.ID.String(),
			Name:        role.Name,
			Description: role.Description,
		}
	}
	return result
}

// mapPermissionsToGraphQL converts repository permissions to GraphQL models
func mapPermissionsToGraphQL(perms []*repository.Permission) []*model.Permission {
	result := make([]*model.Permission, len(perms))
	for i, perm := range perms {
		result[i] = &model.Permission{
			ID:          perm.ID.String(),
			Code:        perm.Code,
			Description: perm.Description,
		}
	}
	return result
}	