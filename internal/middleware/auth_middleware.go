package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/yourusername/oracle-dba-platform/internal/service"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserContextKey is the context key for user information
	UserContextKey contextKey = "user"
)

// UserContext contains authenticated user information
type UserContext struct {
	UserID      uuid.UUID
	Username    string
	Roles       []string
	Permissions []string
}

// AuthMiddleware validates JWT tokens and adds user context
type AuthMiddleware struct {
	authService *service.AuthService
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// Middleware returns an HTTP middleware function
func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Allow unauthenticated requests (will be handled by resolvers)
			next.ServeHTTP(w, r)
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Validate token
		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Parse user ID
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
			return
		}

		// Create user context
		userCtx := &UserContext{
			UserID:      userID,
			Username:    claims.Username,
			Roles:       claims.Roles,
			Permissions: claims.Permissions,
		}

		// Add user context to request context
		ctx := context.WithValue(r.Context(), UserContextKey, userCtx)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext retrieves user context from request context
func GetUserFromContext(ctx context.Context) (*UserContext, bool) {
	user, ok := ctx.Value(UserContextKey).(*UserContext)
	return user, ok
}

// MustGetUserFromContext retrieves user context or panics (use in resolvers after auth check)
func MustGetUserFromContext(ctx context.Context) *UserContext {
	user, ok := GetUserFromContext(ctx)
	if !ok {
		panic("no user in context")
	}
	return user
}

// RequireAuth returns error if user is not authenticated
func RequireAuth(ctx context.Context) (*UserContext, error) {
	user, ok := GetUserFromContext(ctx)
	if !ok {
		return nil, &AuthError{Message: "authentication required"}
	}
	return user, nil
}

// RequirePermission checks if user has a specific permission
func RequirePermission(ctx context.Context, permission string) error {
	user, err := RequireAuth(ctx)
	if err != nil {
		return err
	}

	for _, perm := range user.Permissions {
		if perm == permission {
			return nil
		}
	}

	return &AuthorizationError{
		Message:    "insufficient permissions",
		Permission: permission,
	}
}

// RequireAnyPermission checks if user has any of the specified permissions
func RequireAnyPermission(ctx context.Context, permissions []string) error {
	user, err := RequireAuth(ctx)
	if err != nil {
		return err
	}

	userPerms := make(map[string]bool)
	for _, perm := range user.Permissions {
		userPerms[perm] = true
	}

	for _, permission := range permissions {
		if userPerms[permission] {
			return nil
		}
	}

	return &AuthorizationError{
		Message:    "insufficient permissions",
		Permission: strings.Join(permissions, ", "),
	}
}

// RequireRole checks if user has a specific role
func RequireRole(ctx context.Context, role string) error {
	user, err := RequireAuth(ctx)
	if err != nil {
		return err
	}

	for _, r := range user.Roles {
		if r == role {
			return nil
		}
	}

	return &AuthorizationError{
		Message: "insufficient role",
		Role:    role,
	}
}

// AuthError represents an authentication error
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

// AuthorizationError represents an authorization error
type AuthorizationError struct {
	Message    string
	Permission string
	Role       string
}

func (e *AuthorizationError) Error() string {
	return e.Message
}