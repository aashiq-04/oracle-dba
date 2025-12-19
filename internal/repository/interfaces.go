package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// USER REPOSITORY
// ============================================================================

type User struct {
	ID           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	IsActive     bool
	CreatedAt    time.Time
	LastLogin    *time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
}

// ============================================================================
// ROLE REPOSITORY
// ============================================================================

type Role struct {
	ID          uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
}

type RoleRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
	GetByName(ctx context.Context, name string) (*Role, error)
	List(ctx context.Context) ([]*Role, error)
}

// ============================================================================
// USER ROLES REPOSITORY
// ============================================================================

type UserRole struct {
	UserID     uuid.UUID
	RoleID     uuid.UUID
	AssignedAt time.Time
}

type UserRoleRepository interface {
	Assign(ctx context.Context, userID, roleID uuid.UUID) error
	Revoke(ctx context.Context, userID, roleID uuid.UUID) error
	GetRolesByUserID(ctx context.Context, userID uuid.UUID) ([]*Role, error)
	GetUsersByRoleID(ctx context.Context, roleID uuid.UUID) ([]*User, error)
}

// ============================================================================
// PERMISSION REPOSITORY
// ============================================================================

type Permission struct {
	ID          uuid.UUID
	Code        string
	Description string
}

type PermissionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Permission, error)
	GetByCode(ctx context.Context, code string) (*Permission, error)
	List(ctx context.Context) ([]*Permission, error)
	GetPermissionsByRoleID(ctx context.Context, roleID uuid.UUID) ([]*Permission, error)
	GetPermissionsByUserID(ctx context.Context, userID uuid.UUID) ([]*Permission, error)
}

// ============================================================================
// AUDIT LOG REPOSITORY
// ============================================================================

type AuditLog struct {
	ID              uuid.UUID
	UserID          *uuid.UUID
	Username        string
	Action          string
	ResourceType    string
	ResourceID      *string
	OracleSchema    *string
	Status          string
	IPAddress       *string
	UserAgent       *string
	RequestPayload  *string
	ResponsePayload *string
	ErrorMessage    *string
	DurationMs      *int
	Timestamp       time.Time
}

type AuditLogFilter struct {
	UserID       *uuid.UUID
	Action       *string
	ResourceType *string
	Status       *string
	StartTime    *time.Time
	EndTime      *time.Time
	Limit        int
	Offset       int
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*AuditLog, error)
	List(ctx context.Context, filter *AuditLogFilter) ([]*AuditLog, error)
	Count(ctx context.Context, filter *AuditLogFilter) (int, error)
}

// ============================================================================
// SESSION METRICS REPOSITORY
// ============================================================================

type SessionMetric struct {
	ID               uuid.UUID
	OracleSID        int
	OracleSerial     int
	Username         *string
	SchemaName       *string
	OSUser           *string
	Machine          *string
	Program          *string
	Status           string
	LogonTime        *time.Time
	LastCallET       int
	BlockingSession  *int
	SQLID            *string
	SQLText          *string
	WaitClass        *string
	Event            *string
	SecondsInWait    *int
	CapturedAt       time.Time
}

type SessionMetricsRepository interface {
	Create(ctx context.Context, metrics []*SessionMetric) error
	GetByTimeRange(ctx context.Context, start, end time.Time) ([]*SessionMetric, error)
	GetBySchema(ctx context.Context, schema string, start, end time.Time) ([]*SessionMetric, error)
}

// ============================================================================
// TABLESPACE METRICS REPOSITORY
// ============================================================================

type TablespaceMetric struct {
	ID               uuid.UUID
	TablespaceName   string
	TotalSizeMB      float64
	UsedSizeMB       float64
	FreeSizeMB       float64
	UsagePercentage  float64
	Status           *string
	Contents         *string
	DatafileCount    *int
	CapturedAt       time.Time
}

type TablespaceMetricsRepository interface {
	Create(ctx context.Context, metrics []*TablespaceMetric) error
	GetLatest(ctx context.Context) ([]*TablespaceMetric, error)
	GetByTablespaceName(ctx context.Context, name string, start, end time.Time) ([]*TablespaceMetric, error)
	GetByTimeRange(ctx context.Context, start, end time.Time) ([]*TablespaceMetric, error)
}

// ============================================================================
// QUERY METRICS REPOSITORY
// ============================================================================

type QueryMetric struct {
	ID             uuid.UUID
	SQLID          string
	SQLText        *string
	SchemaName     *string
	ParsingSchema  *string
	Executions     int
	ElapsedTimeMS  float64
	CPUTimeMS      float64
	DiskReads      int
	BufferGets     int
	RowsProcessed  int
	FirstLoadTime  *time.Time
	LastActiveTime *time.Time
	CapturedAt     time.Time
}

type QueryMetricsRepository interface {
	Create(ctx context.Context, metrics []*QueryMetric) error
	GetBySQLID(ctx context.Context, sqlID string, start, end time.Time) ([]*QueryMetric, error)
	GetByTimeRange(ctx context.Context, start, end time.Time) ([]*QueryMetric, error)
}

// ============================================================================
// REPOSITORIES CONTAINER
// ============================================================================

type Repositories struct {
	Users            UserRepository
	Roles            RoleRepository
	UserRoles        UserRoleRepository
	Permissions      PermissionRepository
	AuditLogs        AuditLogRepository
	SessionMetrics   SessionMetricsRepository
	TablespaceMetrics TablespaceMetricsRepository
	QueryMetrics     QueryMetricsRepository
}