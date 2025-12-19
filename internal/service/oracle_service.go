package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/aashiq-04/oracle-dba/internal/repository"
	"github.com/aashiq-04/oracle-dba/pkg/oracle"
)

// OracleService handles Oracle database monitoring operations
type OracleService struct {
	oracleDB              *oracle.OracleDB
	sessionMetricsRepo    repository.SessionMetricsRepository
	tablespaceMetricsRepo repository.TablespaceMetricsRepository
	queryMetricsRepo      repository.QueryMetricsRepository
	auditRepo             repository.AuditLogRepository
}

// NewOracleService creates a new Oracle monitoring service
func NewOracleService(
	oracleDB *oracle.OracleDB,
	sessionMetricsRepo repository.SessionMetricsRepository,
	tablespaceMetricsRepo repository.TablespaceMetricsRepository,
	queryMetricsRepo repository.QueryMetricsRepository,
	auditRepo repository.AuditLogRepository,
) *OracleService {
	return &OracleService{
		oracleDB:              oracleDB,
		sessionMetricsRepo:    sessionMetricsRepo,
		tablespaceMetricsRepo: tablespaceMetricsRepo,
		queryMetricsRepo:      queryMetricsRepo,
		auditRepo:             auditRepo,
	}
}

// ============================================================================
// SESSION MONITORING
// ============================================================================

// OracleSession represents an Oracle database session
type OracleSession struct {
	SID             int
	Serial          int
	Username        *string
	SchemaName      *string
	OSUser          *string
	Machine         *string
	Program         *string
	Status          string
	SQLID           *string
	SQLText         *string
	LogonTime       *time.Time
	LastCallET      int
	BlockingSession *int
	WaitClass       *string
	Event           *string
	SecondsInWait   *int
}

// GetActiveSessions retrieves all active Oracle sessions
func (s *OracleService) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*OracleSession, error) {
	rows, err := s.oracleDB.DB.QueryContext(ctx, oracle.QueryActiveSessions)
	if err != nil {
		s.auditQueryFailure(ctx, userID, "GET_ACTIVE_SESSIONS", err)
		return nil, fmt.Errorf("failed to query active sessions: %w", err)
	}
	defer rows.Close()

	sessions := []*OracleSession{}
	for rows.Next() {
		session := &OracleSession{}
		err := rows.Scan(
			&session.SID,
			&session.Serial,
			&session.Username,
			&session.SchemaName,
			&session.OSUser,
			&session.Machine,
			&session.Program,
			&session.Status,
			&session.SQLID,
			&session.SQLText,
			&session.LogonTime,
			&session.LastCallET,
			&session.BlockingSession,
			&session.WaitClass,
			&session.Event,
			&session.SecondsInWait,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	s.auditQuerySuccess(ctx, userID, "GET_ACTIVE_SESSIONS", len(sessions))
	return sessions, nil
}

// GetAllSessions retrieves all Oracle sessions
func (s *OracleService) GetAllSessions(ctx context.Context, userID uuid.UUID) ([]*OracleSession, error) {
	rows, err := s.oracleDB.DB.QueryContext(ctx, oracle.QueryAllSessions)
	if err != nil {
		s.auditQueryFailure(ctx, userID, "GET_ALL_SESSIONS", err)
		return nil, fmt.Errorf("failed to query all sessions: %w", err)
	}
	defer rows.Close()

	sessions := []*OracleSession{}
	for rows.Next() {
		session := &OracleSession{}
		err := rows.Scan(
			&session.SID,
			&session.Serial,
			&session.Username,
			&session.SchemaName,
			&session.OSUser,
			&session.Machine,
			&session.Program,
			&session.Status,
			&session.SQLID,
			&session.SQLText,
			&session.LogonTime,
			&session.LastCallET,
			&session.BlockingSession,
			&session.WaitClass,
			&session.Event,
			&session.SecondsInWait,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	s.auditQuerySuccess(ctx, userID, "GET_ALL_SESSIONS", len(sessions))
	return sessions, nil
}

// GetSessionsBySchema retrieves sessions for a specific schema
func (s *OracleService) GetSessionsBySchema(ctx context.Context, userID uuid.UUID, schemaName string) ([]*OracleSession, error) {
	rows, err := s.oracleDB.DB.QueryContext(ctx, oracle.QuerySessionsBySchema, schemaName)
	if err != nil {
		s.auditQueryFailure(ctx, userID, "GET_SESSIONS_BY_SCHEMA", err)
		return nil, fmt.Errorf("failed to query sessions by schema: %w", err)
	}
	defer rows.Close()

	sessions := []*OracleSession{}
	for rows.Next() {
		session := &OracleSession{}
		err := rows.Scan(
			&session.SID,
			&session.Serial,
			&session.Username,
			&session.SchemaName,
			&session.OSUser,
			&session.Machine,
			&session.Program,
			&session.Status,
			&session.SQLID,
			&session.SQLText,
			&session.LogonTime,
			&session.LastCallET,
			&session.BlockingSession,
			&session.WaitClass,
			&session.Event,
			&session.SecondsInWait,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	s.auditQuerySuccess(ctx, userID, "GET_SESSIONS_BY_SCHEMA", len(sessions))
	return sessions, nil
}

// ============================================================================
// BLOCKING SESSIONS
// ============================================================================

// BlockingSession represents a blocking session relationship
type BlockingSession struct {
	BlockingSID           int
	BlockingSerial        int
	BlockingUser          *string
	BlockingSchema        *string
	BlockingStatus        string
	BlockingSQLID         *string
	BlockingSQLText       *string
	BlockedSID            int
	BlockedSerial         int
	BlockedUser           *string
	BlockedSchema         *string
	BlockedWaitClass      *string
	BlockedEvent          *string
	BlockedDurationSeconds int
	BlockedSQLText        *string
}

// GetBlockingSessions retrieves all blocking session relationships
func (s *OracleService) GetBlockingSessions(ctx context.Context, userID uuid.UUID) ([]*BlockingSession, error) {
	rows, err := s.oracleDB.DB.QueryContext(ctx, oracle.QueryBlockingSessions)
	if err != nil {
		s.auditQueryFailure(ctx, userID, "GET_BLOCKING_SESSIONS", err)
		return nil, fmt.Errorf("failed to query blocking sessions: %w", err)
	}
	defer rows.Close()

	blockingSessions := []*BlockingSession{}
	for rows.Next() {
		bs := &BlockingSession{}
		err := rows.Scan(
			&bs.BlockingSID,
			&bs.BlockingSerial,
			&bs.BlockingUser,
			&bs.BlockingSchema,
			&bs.BlockingStatus,
			&bs.BlockingSQLID,
			&bs.BlockingSQLText,
			&bs.BlockedSID,
			&bs.BlockedSerial,
			&bs.BlockedUser,
			&bs.BlockedSchema,
			&bs.BlockedWaitClass,
			&bs.BlockedEvent,
			&bs.BlockedDurationSeconds,
			&bs.BlockedSQLText,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan blocking session: %w", err)
		}
		blockingSessions = append(blockingSessions, bs)
	}

	s.auditQuerySuccess(ctx, userID, "GET_BLOCKING_SESSIONS", len(blockingSessions))
	return blockingSessions, nil
}

// ============================================================================
// TABLESPACE MONITORING
// ============================================================================

// Tablespace represents Oracle tablespace information
type Tablespace struct {
	Name            string
	TotalSizeMB     float64
	UsedSizeMB      float64
	FreeSizeMB      float64
	UsagePercentage float64
	Status          string
	Contents        string
	DatafileCount   int
}

// GetTablespaces retrieves all tablespace information
func (s *OracleService) GetTablespaces(ctx context.Context, userID uuid.UUID) ([]*Tablespace, error) {
	rows, err := s.oracleDB.DB.QueryContext(ctx, oracle.QueryTablespaces)
	if err != nil {
		s.auditQueryFailure(ctx, userID, "GET_TABLESPACES", err)
		return nil, fmt.Errorf("failed to query tablespaces: %w", err)
	}
	defer rows.Close()

	tablespaces := []*Tablespace{}
	for rows.Next() {
		ts := &Tablespace{}
		err := rows.Scan(
			&ts.Name,
			&ts.TotalSizeMB,
			&ts.UsedSizeMB,
			&ts.FreeSizeMB,
			&ts.UsagePercentage,
			&ts.Status,
			&ts.Contents,
			&ts.DatafileCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tablespace: %w", err)
		}
		tablespaces = append(tablespaces, ts)
	}

	s.auditQuerySuccess(ctx, userID, "GET_TABLESPACES", len(tablespaces))
	return tablespaces, nil
}

// ============================================================================
// SQL PERFORMANCE MONITORING
// ============================================================================

// SQLPerformance represents SQL query performance metrics
type SQLPerformance struct {
	SQLID            string
	SQLText          *string
	ParsingSchema    *string
	Executions       int
	ElapsedSeconds   float64
	AvgElapsedSeconds float64
	CPUSeconds       float64
	DiskReads        int
	BufferGets       int
	RowsProcessed    int
	FirstLoadTime    *time.Time
	LastActiveTime   *time.Time
}

// GetTopSQLByElapsedTime retrieves top SQL by elapsed time
func (s *OracleService) GetTopSQLByElapsedTime(ctx context.Context, userID uuid.UUID, limit int) ([]*SQLPerformance, error) {
	rows, err := s.oracleDB.DB.QueryContext(ctx, oracle.QueryTopSQLByElapsedTime, limit)
	if err != nil {
		s.auditQueryFailure(ctx, userID, "GET_TOP_SQL_BY_ELAPSED", err)
		return nil, fmt.Errorf("failed to query top SQL by elapsed time: %w", err)
	}
	defer rows.Close()

	sqlPerf := []*SQLPerformance{}
	for rows.Next() {
		sp := &SQLPerformance{}
		err := rows.Scan(
			&sp.SQLID,
			&sp.SQLText,
			&sp.ParsingSchema,
			&sp.Executions,
			&sp.ElapsedSeconds,
			&sp.AvgElapsedSeconds,
			&sp.CPUSeconds,
			&sp.DiskReads,
			&sp.BufferGets,
			&sp.RowsProcessed,
			&sp.FirstLoadTime,
			&sp.LastActiveTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SQL performance: %w", err)
		}
		sqlPerf = append(sqlPerf, sp)
	}

	s.auditQuerySuccess(ctx, userID, "GET_TOP_SQL_BY_ELAPSED", len(sqlPerf))
	return sqlPerf, nil
}

// GetTopSQLByCPU retrieves top SQL by CPU time
func (s *OracleService) GetTopSQLByCPU(ctx context.Context, userID uuid.UUID, limit int) ([]*SQLPerformance, error) {
	rows, err := s.oracleDB.DB.QueryContext(ctx, oracle.QueryTopSQLByCPU, limit)
	if err != nil {
		s.auditQueryFailure(ctx, userID, "GET_TOP_SQL_BY_CPU", err)
		return nil, fmt.Errorf("failed to query top SQL by CPU: %w", err)
	}
	defer rows.Close()

	sqlPerf := []*SQLPerformance{}
	for rows.Next() {
		sp := &SQLPerformance{}
		err := rows.Scan(
			&sp.SQLID,
			&sp.SQLText,
			&sp.ParsingSchema,
			&sp.Executions,
			&sp.CPUSeconds,
			&sp.AvgElapsedSeconds,
			&sp.ElapsedSeconds,
			&sp.DiskReads,
			&sp.BufferGets,
			&sp.RowsProcessed,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SQL performance: %w", err)
		}
		sqlPerf = append(sqlPerf, sp)
	}

	s.auditQuerySuccess(ctx, userID, "GET_TOP_SQL_BY_CPU", len(sqlPerf))
	return sqlPerf, nil
}

// ============================================================================
// DATABASE HEALTH
// ============================================================================

// DatabaseInstance represents Oracle database instance information
type DatabaseInstance struct {
	InstanceName   string
	HostName       string
	Version        string
	StartupTime    time.Time
	Status         string
	DatabaseStatus string
	InstanceRole   string
	UptimeDays     float64
}

// GetDatabaseInstance retrieves database instance information
func (s *OracleService) GetDatabaseInstance(ctx context.Context, userID uuid.UUID) (*DatabaseInstance, error) {
	row := s.oracleDB.DB.QueryRowContext(ctx, oracle.QueryDatabaseInstance)

	instance := &DatabaseInstance{}
	err := row.Scan(
		&instance.InstanceName,
		&instance.HostName,
		&instance.Version,
		&instance.StartupTime,
		&instance.Status,
		&instance.DatabaseStatus,
		&instance.InstanceRole,
		&instance.UptimeDays,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no instance information found")
		}
		s.auditQueryFailure(ctx, userID, "GET_DATABASE_INSTANCE", err)
		return nil, fmt.Errorf("failed to query database instance: %w", err)
	}

	s.auditQuerySuccess(ctx, userID, "GET_DATABASE_INSTANCE", 1)
	return instance, nil
}

// ============================================================================
// SCHEMA MONITORING
// ============================================================================

// SchemaInfo represents schema metadata
type SchemaInfo struct {
	SchemaName      string
	TotalObjects    int
	TableCount      int
	IndexCount      int
	ViewCount       int
	ProcedureCount  int
	FunctionCount   int
	PackageCount    int
}

// GetSchemas retrieves all schemas with object counts
func (s *OracleService) GetSchemas(ctx context.Context, userID uuid.UUID) ([]*SchemaInfo, error) {
	rows, err := s.oracleDB.DB.QueryContext(ctx, oracle.QuerySchemas)
	if err != nil {
		s.auditQueryFailure(ctx, userID, "GET_SCHEMAS", err)
		return nil, fmt.Errorf("failed to query schemas: %w", err)
	}
	defer rows.Close()

	schemas := []*SchemaInfo{}
	for rows.Next() {
		schema := &SchemaInfo{}
		err := rows.Scan(
			&schema.SchemaName,
			&schema.TotalObjects,
			&schema.TableCount,
			&schema.IndexCount,
			&schema.ViewCount,
			&schema.ProcedureCount,
			&schema.FunctionCount,
			&schema.PackageCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}
		schemas = append(schemas, schema)
	}

	s.auditQuerySuccess(ctx, userID, "GET_SCHEMAS", len(schemas))
	return schemas, nil
}

// ============================================================================
// AUDIT HELPERS
// ============================================================================

func (s *OracleService) auditQuerySuccess(ctx context.Context, userID uuid.UUID, action string, count int) {
	resourceID := fmt.Sprintf("count:%d", count)
	log := &repository.AuditLog{
		UserID:       &userID,
		Username:     userID.String(),
		Action:       action,
		ResourceType: "ORACLE_QUERY",
		ResourceID:   &resourceID,
		Status:       "SUCCESS",
	}
	_ = s.auditRepo.Create(ctx, log)
}

func (s *OracleService) auditQueryFailure(ctx context.Context, userID uuid.UUID, action string, err error) {
	errMsg := err.Error()
	log := &repository.AuditLog{
		UserID:       &userID,
		Username:     userID.String(),
		Action:       action,
		ResourceType: "ORACLE_QUERY",
		Status:       "FAILURE",
		ErrorMessage: &errMsg,
	}
	_ = s.auditRepo.Create(ctx, log)
}