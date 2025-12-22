package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/aashiq-04/oracle-dba/internal/repository"
)

// RBACService handles role-based access control operations
type RBACService struct {
	userRoleRepo   repository.UserRoleRepository
	permissionRepo repository.PermissionRepository
	roleRepo       repository.RoleRepository
	auditRepo      repository.AuditLogRepository
}

// NewRBACService creates a new RBAC service
func NewRBACService(
	userRoleRepo repository.UserRoleRepository,
	permissionRepo repository.PermissionRepository,
	roleRepo repository.RoleRepository,
	auditRepo repository.AuditLogRepository,
) *RBACService {
	return &RBACService{
		userRoleRepo:   userRoleRepo,
		permissionRepo: permissionRepo,
		roleRepo:       roleRepo,
		auditRepo:      auditRepo,
	}
}

// HasPermission checks if a user has a specific permission
func (s *RBACService) HasPermission(ctx context.Context, userID uuid.UUID, permissionCode string) (bool, error) {
	permissions, err := s.permissionRepo.GetPermissionsByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user permissions: %w", err)
	}

	for _, perm := range permissions {
		if perm.Code == permissionCode {
			return true, nil
		}
	}

	return false, nil
}

// HasAnyPermission checks if a user has any of the specified permissions
func (s *RBACService) HasAnyPermission(ctx context.Context, userID uuid.UUID, permissionCodes []string) (bool, error) {
	permissions, err := s.permissionRepo.GetPermissionsByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user permissions: %w", err)
	}

	permMap := make(map[string]bool)
	for _, perm := range permissions {
		permMap[perm.Code] = true
	}

	for _, code := range permissionCodes {
		if permMap[code] {
			return true, nil
		}
	}

	return false, nil
}

// HasAllPermissions checks if a user has all of the specified permissions
func (s *RBACService) HasAllPermissions(ctx context.Context, userID uuid.UUID, permissionCodes []string) (bool, error) {
	permissions, err := s.permissionRepo.GetPermissionsByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user permissions: %w", err)
	}

	permMap := make(map[string]bool)
	for _, perm := range permissions {
		permMap[perm.Code] = true
	}

	for _, code := range permissionCodes {
		if !permMap[code] {
			return false, nil
		}
	}

	return true, nil
}

// HasRole checks if a user has a specific role
func (s *RBACService) HasRole(ctx context.Context, userID uuid.UUID, roleName string) (bool, error) {
	roles, err := s.userRoleRepo.GetRolesByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user roles: %w", err)
	}

	for _, role := range roles {
		if role.Name == roleName {
			return true, nil
		}
	}

	return false, nil
}

// GetUserRoles returns all roles for a user
func (s *RBACService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*repository.Role, error) {
	return s.userRoleRepo.GetRolesByUserID(ctx, userID)
}

// GetUserPermissions returns all permissions for a user
func (s *RBACService) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*repository.Permission, error) {
	return s.permissionRepo.GetPermissionsByUserID(ctx, userID)
}

// AssignRole assigns a role to a user
func (s *RBACService) AssignRole(ctx context.Context, userID, roleID uuid.UUID, assignedBy uuid.UUID) error {
	if err := s.userRoleRepo.Assign(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	// Audit role assignment
	s.auditRoleAssignment(ctx, userID, roleID, assignedBy)

	return nil
}

// RevokeRole revokes a role from a user
func (s *RBACService) RevokeRole(ctx context.Context, userID, roleID uuid.UUID, revokedBy uuid.UUID) error {
	if err := s.userRoleRepo.Revoke(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	// Audit role revocation
	s.auditRoleRevocation(ctx, userID, roleID, revokedBy)

	return nil
}

// GetAllRoles returns all available roles
func (s *RBACService) GetAllRoles(ctx context.Context) ([]*repository.Role, error) {
	return s.roleRepo.List(ctx)
}

// GetAllPermissions returns all available permissions
func (s *RBACService) GetAllPermissions(ctx context.Context) ([]*repository.Permission, error) {
	return s.permissionRepo.List(ctx)
}

// GetRolePermissions returns all permissions for a role
func (s *RBACService) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*repository.Permission, error) {
	return s.permissionRepo.GetPermissionsByRoleID(ctx, roleID)
}

// CheckAccess validates if a user can perform an action on a resource
// This is the main RBAC enforcement point
func (s *RBACService) CheckAccess(ctx context.Context, userID uuid.UUID, requiredPermission string) error {
	hasPermission, err := s.HasPermission(ctx, userID, requiredPermission)
	if err != nil {
		return fmt.Errorf("failed to check permission: %w", err)
	}

	if !hasPermission {
		// Audit access denied
		s.auditAccessDenied(ctx, userID, requiredPermission)
		return fmt.Errorf("access denied: missing permission '%s'", requiredPermission)
	}

	return nil
}

// Audit helper functions

func (s *RBACService) auditRoleAssignment(ctx context.Context, userID, roleID, assignedBy uuid.UUID) {
	resourceID := roleID.String()
	log := &repository.AuditLog{
		UserID:       &assignedBy,
		Username:     assignedBy.String(), // TODO: resolve username
		Action:       "ASSIGN_ROLE",
		ResourceType: "USER_ROLE",
		ResourceID:   &resourceID,
		Status:       "SUCCESS",
	}
	_ = s.auditRepo.Create(ctx, log)
}

func (s *RBACService) auditRoleRevocation(ctx context.Context, userID, roleID, revokedBy uuid.UUID) {
	resourceID := roleID.String()
	log := &repository.AuditLog{
		UserID:       &revokedBy,
		Username:     revokedBy.String(), // TODO: resolve username
		Action:       "REVOKE_ROLE",
		ResourceType: "USER_ROLE",
		ResourceID:   &resourceID,
		Status:       "SUCCESS",
	}
	_ = s.auditRepo.Create(ctx, log)
}

func (s *RBACService) auditAccessDenied(ctx context.Context, userID uuid.UUID, permission string) {
	log := &repository.AuditLog{
		UserID:       &userID,
		Username:     userID.String(), // TODO: resolve username
		Action:       "ACCESS_DENIED",
		ResourceType: "PERMISSION",
		Status:       "DENIED",
		ErrorMessage: &permission,
	}
	_ = s.auditRepo.Create(ctx, log)
}
