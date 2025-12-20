package graph

import (
	"context"
	"fmt"

	"github.com/aashiq-04/oracle-dba/internal/graph/model"
	"github.com/aashiq-04/oracle-dba/internal/middleware"
	"github.com/aashiq-04/oracle-dba/internal/service"
)

// ============================================================================
// QUERY RESOLVERS - Oracle Session Monitoring
// ============================================================================

// Sessions returns all Oracle sessions
func (r *queryResolver) Sessions(ctx context.Context, filter *model.SessionFilterInput) ([]*model.OracleSession, error) {
	// Require VIEW_SESSIONS permission
	if err := middleware.RequirePermission(ctx, "VIEW_SESSIONS"); err != nil {
		return nil, err
	}

	userCtx := middleware.MustGetUserFromContext(ctx)

	// Get sessions from Oracle
	var sessions []*service.OracleSession
	var err error

	if filter != nil && filter.SchemaName != nil {
		// Filter by schema
		sessions, err = r.oracleService.GetSessionsBySchema(ctx, userCtx.UserID, *filter.SchemaName)
	} else {
		// Get all sessions
		sessions, err = r.oracleService.GetAllSessions(ctx, userCtx.UserID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	// Filter by status if provided
	if filter != nil && filter.Status != nil {
		filtered := []*service.OracleSession{}
		for _, session := range sessions {
			if session.Status == string(*filter.Status) {
				filtered = append(filtered, session)
			}
		}
		sessions = filtered
	}

	return mapSessionsToGraphQL(sessions), nil
}

// ActiveSessions returns only active Oracle sessions
func (r *queryResolver) ActiveSessions(ctx context.Context, filter *model.SessionFilterInput) ([]*model.OracleSession, error) {
	// Require VIEW_SESSIONS permission
	if err := middleware.RequirePermission(ctx, "VIEW_SESSIONS"); err != nil {
		return nil, err
	}

	userCtx := middleware.MustGetUserFromContext(ctx)

	// Get active sessions from Oracle
	sessions, err := r.oracleService.GetActiveSessions(ctx, userCtx.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}

	return mapSessionsToGraphQL(sessions), nil
}

// SessionSummary returns session statistics
func (r *queryResolver) SessionSummary(ctx context.Context) (*model.SessionSummary, error) {
	// Require VIEW_SESSIONS permission
	if err := middleware.RequirePermission(ctx, "VIEW_SESSIONS"); err != nil {
		return nil, err
	}

	userCtx := middleware.MustGetUserFromContext(ctx)

	// Get all sessions
	sessions, err := r.oracleService.GetAllSessions(ctx, userCtx.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	// Calculate summary
	summary := &model.SessionSummary{
		TotalSessions:   len(sessions),
		ActiveSessions:  0,
		InactiveSessions: 0,
		BlockedSessions: 0,
		BySchema:        []*model.SessionsBySchema{},
	}

	schemaMap := make(map[string]*model.SessionsBySchema)

	for _, session := range sessions {
		// Count by status
		if session.Status == "ACTIVE" {
			summary.ActiveSessions++
		} else {
			summary.InactiveSessions++
		}

		// Count blocked sessions
		if session.BlockingSession != nil {
			summary.BlockedSessions++
		}

		// Group by schema
		if session.SchemaName != nil {
			schemaName := *session.SchemaName
			if _, exists := schemaMap[schemaName]; !exists {
				schemaMap[schemaName] = &model.SessionsBySchema{
					SchemaName: schemaName,
					Total:      0,
					Active:     0,
					Inactive:   0,
				}
			}

			schemaMap[schemaName].Total++
			if session.Status == "ACTIVE" {
				schemaMap[schemaName].Active++
			} else {
				schemaMap[schemaName].Inactive++
			}
		}
	}

	// Convert schema map to slice
	for _, schemaStats := range schemaMap {
		summary.BySchema = append(summary.BySchema, schemaStats)
	}

	return summary, nil
}

// Session returns a specific session by SID
func (r *queryResolver) Session(ctx context.Context, sid int) (*model.OracleSession, error) {
	// Require VIEW_SESSIONS permission
	if err := middleware.RequirePermission(ctx, "VIEW_SESSIONS"); err != nil {
		return nil, err
	}

	// Implementation would filter sessions by SID
	return nil, fmt.Errorf("not implemented")
}

// ============================================================================
// QUERY RESOLVERS - Lock Monitoring
// ============================================================================

// BlockingSessions returns all blocking session relationships
func (r *queryResolver) BlockingSessions(ctx context.Context) ([]*model.BlockingSession, error) {
	// Require VIEW_LOCKS permission
	if err := middleware.RequirePermission(ctx, "VIEW_LOCKS"); err != nil {
		return nil, err
	}

	userCtx := middleware.MustGetUserFromContext(ctx)

	// Get blocking sessions from Oracle
	blockingSessions, err := r.oracleService.GetBlockingSessions(ctx, userCtx.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocking sessions: %w", err)
	}

	return mapBlockingSessionsToGraphQL(blockingSessions), nil
}

// Locks returns lock information (not yet fully implemented)
func (r *queryResolver) Locks(ctx context.Context, schemaName *string) ([]*model.LockInfo, error) {
	// Require VIEW_LOCKS permission
	if err := middleware.RequirePermission(ctx, "VIEW_LOCKS"); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("not implemented")
}

// ============================================================================
// QUERY RESOLVERS - Tablespace Monitoring
// ============================================================================

// Tablespaces returns all tablespace information
func (r *queryResolver) Tablespaces(ctx context.Context, filter *model.TablespaceFilterInput) ([]*model.Tablespace, error) {
	// Require VIEW_TABLESPACES permission
	if err := middleware.RequirePermission(ctx, "VIEW_TABLESPACES"); err != nil {
		return nil, err
	}

	userCtx := middleware.MustGetUserFromContext(ctx)

	// Get tablespaces from Oracle
	tablespaces, err := r.oracleService.GetTablespaces(ctx, userCtx.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tablespaces: %w", err)
	}

	// Apply filters if provided
	if filter != nil {
		filtered := []*service.Tablespace{}
		for _, ts := range tablespaces {
			// Filter by name
			if filter.Name != nil && ts.Name != *filter.Name {
				continue
			}
			// Filter by minimum usage percentage
			if filter.MinUsagePercentage != nil && ts.UsagePercentage < *filter.MinUsagePercentage {
				continue
			}
			filtered = append(filtered, ts)
		}
		tablespaces = filtered
	}

	return mapTablespacesToGraphQL(tablespaces), nil
}

// Tablespace returns a specific tablespace by name
func (r *queryResolver) Tablespace(ctx context.Context, name string) (*model.Tablespace, error) {
	// Require VIEW_TABLESPACES permission
	if err := middleware.RequirePermission(ctx, "VIEW_TABLESPACES"); err != nil {
		return nil, err
	}

	userCtx := middleware.MustGetUserFromContext(ctx)

	// Get tablespaces and filter
	tablespaces, err := r.oracleService.GetTablespaces(ctx, userCtx.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tablespaces: %w", err)
	}

	for _, ts := range tablespaces {
		if ts.Name == name {
			return mapTablespaceToGraphQL(ts), nil
		}
	}

	return nil, fmt.Errorf("tablespace not found")
}

// TablespaceHistory placeholder
func (r *queryResolver) TablespaceHistory(ctx context.Context, name string, timeRange model.TimeRangeInput) ([]*model.TablespaceMetric, error) {
	if err := middleware.RequirePermission(ctx, "VIEW_TABLESPACES"); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("not implemented")
}

// TablespaceGrowth placeholder
func (r *queryResolver) TablespaceGrowth(ctx context.Context, name string, days int) (*model.TablespaceGrowth, error) {
	if err := middleware.RequirePermission(ctx, "VIEW_TABLESPACES"); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("not implemented")
}

// ============================================================================
// QUERY RESOLVERS - SQL Performance
// ============================================================================

// TopSQLByElapsedTime returns top SQL queries by elapsed time
func (r *queryResolver) TopSqlByElapsedTime(ctx context.Context, limit int) ([]*model.SqlPerformance, error) {
	// Require VIEW_SQL permission
	if err := middleware.RequirePermission(ctx, "VIEW_SQL"); err != nil {
		return nil, err
	}

	userCtx := middleware.MustGetUserFromContext(ctx)

	// Get top SQL from Oracle
	sqlPerf, err := r.oracleService.GetTopSQLByElapsedTime(ctx, userCtx.UserID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top SQL by elapsed time: %w", err)
	}

	return mapSQLPerformanceToGraphQL(sqlPerf), nil
}

// TopSQLByCpuTime returns top SQL queries by CPU time
func (r *queryResolver) TopSqlByCpuTime(ctx context.Context, limit int) ([]*model.SqlPerformance, error) {
	// Require VIEW_SQL permission
	if err := middleware.RequirePermission(ctx, "VIEW_SQL"); err != nil {
		return nil, err
	}

	userCtx := middleware.MustGetUserFromContext(ctx)

	// Get top SQL from Oracle
	sqlPerf, err := r.oracleService.GetTopSQLByCPU(ctx, userCtx.UserID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top SQL by CPU time: %w", err)
	}

	return mapSQLPerformanceToGraphQL(sqlPerf), nil
}

// Placeholders for other SQL performance queries
func (r *queryResolver) TopSqlByExecutions(ctx context.Context, limit int) ([]*model.SqlPerformance, error) {
	if err := middleware.RequirePermission(ctx, "VIEW_SQL"); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("not implemented")
}

func (r *queryResolver) TopSqlByDiskReads(ctx context.Context, limit int) ([]*model.SqlPerformance, error) {
	if err := middleware.RequirePermission(ctx, "VIEW_SQL"); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("not implemented")
}

func (r *queryResolver) SqlPerformance(ctx context.Context, filter *model.SqlPerformanceFilterInput) ([]*model.SqlPerformance, error) {
	if err := middleware.RequirePermission(ctx, "VIEW_SQL"); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("not implemented")
}

func (r *queryResolver) SqlByID(ctx context.Context, sqlID string) (*model.SqlPerformance, error) {
	if err := middleware.RequirePermission(ctx, "VIEW_SQL"); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("not implemented")
}

func (r *queryResolver) SqlHistory(ctx context.Context, sqlID string, timeRange model.TimeRangeInput) ([]*model.SqlMetric, error) {
	if err := middleware.RequirePermission(ctx, "VIEW_SQL"); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("not implemented")
}

// ============================================================================
// QUERY RESOLVERS - Schema Monitoring
// ============================================================================

// Schemas returns all schemas with object counts
func (r *queryResolver) Schemas(ctx context.Context) ([]*model.SchemaInfo, error) {
	// Require VIEW_SCHEMA permission
	if err := middleware.RequirePermission(ctx, "VIEW_SCHEMA"); err != nil {
		return nil, err
	}

	userCtx := middleware.MustGetUserFromContext(ctx)

	// Get schemas from Oracle
	schemas, err := r.oracleService.GetSchemas(ctx, userCtx.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schemas: %w", err)
	}

	return mapSchemasToGraphQL(schemas), nil
}

// Placeholders for other schema queries
func (r *queryResolver) SchemaInfo(ctx context.Context, name string) (*model.SchemaInfo, error) {
	if err := middleware.RequirePermission(ctx, "VIEW_SCHEMA"); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("not implemented")
}

func (r *queryResolver) InvalidObjects(ctx context.Context, schemaName *string) ([]*model.InvalidObject, error) {
	if err := middleware.RequirePermission(ctx, "VIEW_SCHEMA"); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("not implemented")
}

func (r *queryResolver) RecentSchemaChanges(ctx context.Context, schemaName *string, days int) ([]*model.SchemaChange, error) {
	if err := middleware.RequirePermission(ctx, "VIEW_SCHEMA"); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("not implemented")
}

// ============================================================================
// QUERY RESOLVERS - Database Health
// ============================================================================

// DatabaseInstance returns database instance information
func (r *queryResolver) DatabaseInstance(ctx context.Context) (*model.DatabaseInstance, error) {
	// Require authentication
	if _, err := middleware.RequireAuth(ctx); err != nil {
		return nil, err
	}

	userCtx := middleware.MustGetUserFromContext(ctx)

	// Get database instance info from Oracle
	instance, err := r.oracleService.GetDatabaseInstance(ctx, userCtx.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	return mapDatabaseInstanceToGraphQL(instance), nil
}

// DatabaseSize placeholder
func (r *queryResolver) DatabaseSize(ctx context.Context) (*model.DatabaseSize, error) {
	if _, err := middleware.RequireAuth(ctx); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("not implemented")
}

// ============================================================================
// QUERY RESOLVERS - Audit Logs
// ============================================================================

func (r *queryResolver) AuditLogs(ctx context.Context, filter *model.AuditLogFilterInput, limit int, offset int) ([]*model.AuditLog, error) {
	// Require AUDIT_READ permission
	if err := middleware.RequirePermission(ctx, "AUDIT_READ"); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("not implemented")
}

func (r *queryResolver) AuditLog(ctx context.Context, id string) (*model.AuditLog, error) {
	// Require AUDIT_READ permission
	if err := middleware.RequirePermission(ctx, "AUDIT_READ"); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("not implemented")
}

// ============================================================================
// MUTATION RESOLVERS - Session Management
// ============================================================================

// KillSession kills an Oracle session (DBA only)
func (r *mutationResolver) KillSession(ctx context.Context, sid int, serial int) (bool, error) {
	// Require SESSION_KILL permission (only DBA/Admin should have this)
	if err := middleware.RequirePermission(ctx, "SESSION_KILL"); err != nil {
		return false, err
	}

	// Implementation would execute ALTER SYSTEM KILL SESSION
	return false, fmt.Errorf("not implemented - would execute: ALTER SYSTEM KILL SESSION '%d,%d'", sid, serial)
}

// ============================================================================
// HELPER FUNCTIONS - Mappers
// ============================================================================

func mapSessionsToGraphQL(sessions []*service.OracleSession) []*model.OracleSession {
	result := make([]*model.OracleSession, len(sessions))
	for i, session := range sessions {
		status := model.SessionStatus(session.Status)
		result[i] = &model.OracleSession{
			Sid:             session.SID,
			Serial:          session.Serial,
			Username:        session.Username,
			SchemaName:      session.SchemaName,
			OsUser:          session.OSUser,
			Machine:         session.Machine,
			Program:         session.Program,
			Status:          status,
			SqlID:           session.SQLID,
			SqlText:         session.SQLText,
			LogonTime:       session.LogonTime,
			LastCallSeconds: session.LastCallET,
			BlockingSession: session.BlockingSession,
			WaitClass:       session.WaitClass,
			Event:           session.Event,
			SecondsInWait:   session.SecondsInWait,
		}
	}
	return result
}

func mapBlockingSessionsToGraphQL(sessions []*service.BlockingSession) []*model.BlockingSession {
	result := make([]*model.BlockingSession, len(sessions))
	for i, bs := range sessions {
		blockingStatus := model.SessionStatus(bs.BlockingStatus)
		result[i] = &model.BlockingSession{
			BlockingSid:            bs.BlockingSID,
			BlockingSerial:         bs.BlockingSerial,
			BlockingUser:           bs.BlockingUser,
			BlockingSchema:         bs.BlockingSchema,
			BlockingStatus:         blockingStatus,
			BlockingSqlID:          bs.BlockingSQLID,
			BlockingSqlText:        bs.BlockingSQLText,
			BlockedSid:             bs.BlockedSID,
			BlockedSerial:          bs.BlockedSerial,
			BlockedUser:            bs.BlockedUser,
			BlockedSchema:          bs.BlockedSchema,
			BlockedWaitClass:       bs.BlockedWaitClass,
			BlockedEvent:           bs.BlockedEvent,
			BlockedDurationSeconds: bs.BlockedDurationSeconds,
			BlockedSqlText:         bs.BlockedSQLText,
		}
	}
	return result
}

func mapTablespacesToGraphQL(tablespaces []*service.Tablespace) []*model.Tablespace {
	result := make([]*model.Tablespace, len(tablespaces))
	for i, ts := range tablespaces {
		contents := model.TablespaceContents(ts.Contents)
		result[i] = &model.Tablespace{
			Name:            ts.Name,
			TotalSizeMb:     ts.TotalSizeMB,
			UsedSizeMb:      ts.UsedSizeMB,
			FreeSizeMb:      ts.FreeSizeMB,
			UsagePercentage: ts.UsagePercentage,
			Status:          ts.Status,
			Contents:        contents,
			DatafileCount:   ts.DatafileCount,
		}
	}
	return result
}

func mapTablespaceToGraphQL(ts *service.Tablespace) *model.Tablespace {
	contents := model.TablespaceContents(ts.Contents)
	return &model.Tablespace{
		Name:            ts.Name,
		TotalSizeMb:     ts.TotalSizeMB,
		UsedSizeMb:      ts.UsedSizeMB,
		FreeSizeMb:      ts.FreeSizeMB,
		UsagePercentage: ts.UsagePercentage,
		Status:          ts.Status,
		Contents:        contents,
		DatafileCount:   ts.DatafileCount,
	}
}

func mapSQLPerformanceToGraphQL(sqlPerf []*service.SQLPerformance) []*model.SqlPerformance {
	result := make([]*model.SqlPerformance, len(sqlPerf))
	for i, sp := range sqlPerf {
		result[i] = &model.SqlPerformance{
			SqlID:           sp.SQLID,
			SqlText:         sp.SQLText,
			SchemaName:      sp.ParsingSchema,
			ParsingSchema:   sp.ParsingSchema,
			Executions:      sp.Executions,
			ElapsedTimeMs:   sp.ElapsedSeconds * 1000, // Convert to milliseconds
			AvgElapsedMs:    sp.AvgElapsedSeconds * 1000,
			CpuTimeMs:       sp.CPUSeconds * 1000,
			AvgCpuMs:        0, // Would need to calculate
			DiskReads:       sp.DiskReads,
			BufferGets:      sp.BufferGets,
			RowsProcessed:   sp.RowsProcessed,
			FirstLoadTime:   sp.FirstLoadTime,
			LastActiveTime:  sp.LastActiveTime,
		}
	}
	return result
}

func mapSchemasToGraphQL(schemas []*service.SchemaInfo) []*model.SchemaInfo {
	result := make([]*model.SchemaInfo, len(schemas))
	for i, schema := range schemas {
		result[i] = &model.SchemaInfo{
			SchemaName:     schema.SchemaName,
			TotalObjects:   schema.TotalObjects,
			TableCount:     schema.TableCount,
			IndexCount:     schema.IndexCount,
			ViewCount:      schema.ViewCount,
			ProcedureCount: schema.ProcedureCount,
			FunctionCount:  schema.FunctionCount,
			PackageCount:   schema.PackageCount,
		}
	}
	return result
}

func mapDatabaseInstanceToGraphQL(instance *service.DatabaseInstance) *model.DatabaseInstance {
	return &model.DatabaseInstance{
		InstanceName:   instance.InstanceName,
		HostName:       instance.HostName,
		Version:        instance.Version,
		StartupTime:    instance.StartupTime,
		Status:         instance.Status,
		DatabaseStatus: instance.DatabaseStatus,
		InstanceRole:   instance.InstanceRole,
		UptimeDays:     instance.UptimeDays,
	}
}