package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/aashiq-04/oracle-dba/internal/repository"
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo       repository.UserRepository
	userRoleRepo   repository.UserRoleRepository
	permissionRepo repository.PermissionRepository
	auditRepo      repository.AuditLogRepository
	jwtSecret      string
	jwtExpiration  time.Duration
	jwtIssuer      string
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo repository.UserRepository,
	userRoleRepo repository.UserRoleRepository,
	permissionRepo repository.PermissionRepository,
	auditRepo repository.AuditLogRepository,
	jwtSecret string,
	jwtExpiration time.Duration,
	jwtIssuer string,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		userRoleRepo:   userRoleRepo,
		permissionRepo: permissionRepo,
		auditRepo:      auditRepo,
		jwtSecret:      jwtSecret,
		jwtExpiration:  jwtExpiration,
		jwtIssuer:      jwtIssuer,
	}
}

// LoginResponse contains the authentication token and user info
type LoginResponse struct {
	Token     string
	User      *repository.User
	Roles     []*repository.Role
	ExpiresAt time.Time
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		// Log failed login attempt
		s.auditFailedLogin(ctx, username, "user not found")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		s.auditFailedLogin(ctx, username, "user inactive")
		return nil, fmt.Errorf("user account is inactive")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.auditFailedLogin(ctx, username, "invalid password")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Get user roles
	roles, err := s.userRoleRepo.GetRolesByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Get user permissions
	permissions, err := s.permissionRepo.GetPermissionsByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	// Generate JWT token
	expiresAt := time.Now().Add(s.jwtExpiration)
	token, err := s.generateJWT(user.ID, user.Username, roles, permissions, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Update last login time
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail login
		fmt.Printf("failed to update last login: %v\n", err)
	}

	// Audit successful login
	s.auditSuccessfulLogin(ctx, user.ID, username)

	return &LoginResponse{
		Token:     token,
		User:      user,
		Roles:     roles,
		ExpiresAt: expiresAt,
	}, nil
}

// CreateUser creates a new user with hashed password
func (s *AuthService) CreateUser(ctx context.Context, username, email, password string, roleIDs []uuid.UUID) (*repository.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &repository.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Assign roles
	for _, roleID := range roleIDs {
		if err := s.userRoleRepo.Assign(ctx, user.ID, roleID); err != nil {
			return nil, fmt.Errorf("failed to assign role: %w", err)
		}
	}

	// Audit user creation
	s.auditUserCreation(ctx, user.ID, username)

	return user, nil
}

// GetUserByID gets a user by their ID
func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*repository.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
// UpdateUser updates user information
func (s *AuthService) UpdateUser(ctx context.Context, userID uuid.UUID, email string, isActive bool) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	user.Email = email
	user.IsActive = isActive

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = string(hashedPassword)

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// ValidateToken validates a JWT token and returns claims
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// generateJWT generates a JWT token
func (s *AuthService) generateJWT(
	userID uuid.UUID,
	username string,
	roles []*repository.Role,
	permissions []*repository.Permission,
	expiresAt time.Time,
) (string, error) {
	// Extract role names
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	// Extract permission codes
	permCodes := make([]string, len(permissions))
	for i, perm := range permissions {
		permCodes[i] = perm.Code
	}

	// Create claims
	claims := JWTClaims{
		UserID:      userID.String(),
		Username:    username,
		Roles:       roleNames,
		Permissions: permCodes,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.jwtIssuer,
			Subject:   username,
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// Audit helper functions

func (s *AuthService) auditSuccessfulLogin(ctx context.Context, userID uuid.UUID, username string) {
	log := &repository.AuditLog{
		UserID:       &userID,
		Username:     username,
		Action:       "LOGIN",
		ResourceType: "AUTH",
		Status:       "SUCCESS",
	}
	_ = s.auditRepo.Create(ctx, log)
}

func (s *AuthService) auditFailedLogin(ctx context.Context, username, reason string) {
	log := &repository.AuditLog{
		Username:     username,
		Action:       "LOGIN",
		ResourceType: "AUTH",
		Status:       "FAILURE",
		ErrorMessage: &reason,
	}
	_ = s.auditRepo.Create(ctx, log)
}

func (s *AuthService) auditUserCreation(ctx context.Context, userID uuid.UUID, username string) {
	resourceID := userID.String()
	log := &repository.AuditLog{
		UserID:       &userID,
		Username:     username,
		Action:       "CREATE_USER",
		ResourceType: "USER",
		ResourceID:   &resourceID,
		Status:       "SUCCESS",
	}
	_ = s.auditRepo.Create(ctx, log)
}
