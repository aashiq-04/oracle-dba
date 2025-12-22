package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	go_ora "github.com/sijms/go-ora/v2"
)

// OracleDB wraps the Oracle connection pool
type OracleDB struct {
	DB *sql.DB
}

// OracleConfig holds Oracle connection configuration
type OracleConfig struct {
	Host        string
	Port        string
	ServiceName string
	Username    string
	Password    string
	MaxConns    int
	MinConns    int
}

// NewOracleDB creates a new Oracle connection pool
func NewOracleDB(cfg OracleConfig) (*OracleDB, error) {
	// Convert port string to int
	portInt, err := strconv.Atoi(cfg.Port)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	// Pure Go Oracle driver connection string
	dsn := go_ora.BuildUrl(cfg.Host, portInt, cfg.ServiceName, cfg.Username, cfg.Password, nil)

	db, err := sql.Open("oracle", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open oracle connection: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.MaxConns)
	db.SetMaxIdleConns(cfg.MinConns)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(10 * time.Minute)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping oracle: %w", err)
	}

	return &OracleDB{DB: db}, nil
}

// Close closes the Oracle connection pool
func (o *OracleDB) Close() error {
	return o.DB.Close()
}

// Health checks the health of the Oracle connection
func (o *OracleDB) Health(ctx context.Context) error {
	return o.DB.PingContext(ctx)
}

// Stats returns connection pool statistics
func (o *OracleDB) Stats() sql.DBStats {
	return o.DB.Stats()
}

// Rest of the file remains the same (all the query constants)
// ... (keep all the QueryActiveSessions, QueryBlockingSessions, etc.)
// ============================================================================
// ORACLE MONITORING QUERIES (CONSTANTS)
// ============================================================================

const (
	// QueryActiveSessions retrieves all active Oracle sessions
	QueryActiveSessions = `
		SELECT
			s.sid,
			s.serial#,
			s.username,
			s.schemaname,
			s.osuser,
			s.machine,
			s.program,
			s.status,
			s.sql_id,
			sq.sql_text,
			s.logon_time,
			s.last_call_et,
			s.blocking_session,
			s.wait_class,
			s.event,
			s.seconds_in_wait
		FROM v$session s
		LEFT JOIN v$sql sq ON s.sql_id = sq.sql_id
		WHERE s.type = 'USER'
		  AND s.username IS NOT NULL
		  AND s.status = 'ACTIVE'
		ORDER BY s.last_call_et DESC
	`

	// QueryAllSessions retrieves all user sessions
	QueryAllSessions = `
		SELECT
			s.sid,
			s.serial#,
			s.username,
			s.schemaname,
			s.osuser,
			s.machine,
			s.program,
			s.status,
			s.sql_id,
			sq.sql_text,
			s.logon_time,
			s.last_call_et,
			s.blocking_session,
			s.wait_class,
			s.event,
			s.seconds_in_wait
		FROM v$session s
		LEFT JOIN v$sql sq ON s.sql_id = sq.sql_id
		WHERE s.type = 'USER'
		  AND s.username IS NOT NULL
		ORDER BY s.logon_time DESC
	`

	// QuerySessionsBySchema retrieves sessions for a specific schema
	QuerySessionsBySchema = `
		SELECT
			s.sid,
			s.serial#,
			s.username,
			s.schemaname,
			s.osuser,
			s.machine,
			s.program,
			s.status,
			s.sql_id,
			sq.sql_text,
			s.logon_time,
			s.last_call_et,
			s.blocking_session,
			s.wait_class,
			s.event,
			s.seconds_in_wait
		FROM v$session s
		LEFT JOIN v$sql sq ON s.sql_id = sq.sql_id
		WHERE s.type = 'USER'
		  AND s.username IS NOT NULL
		  AND s.schemaname = :1
		ORDER BY s.last_call_et DESC
	`

	// QueryBlockingSessions retrieves blocking session information
	QueryBlockingSessions = `
		SELECT
			blocking.sid as blocking_sid,
			blocking.serial# as blocking_serial,
			blocking.username as blocking_user,
			blocking.schemaname as blocking_schema,
			blocking.status as blocking_status,
			blocking.sql_id as blocking_sql_id,
			blocking_sql.sql_text as blocking_sql_text,
			blocked.sid as blocked_sid,
			blocked.serial# as blocked_serial,
			blocked.username as blocked_user,
			blocked.schemaname as blocked_schema,
			blocked.wait_class as blocked_wait_class,
			blocked.event as blocked_event,
			blocked.seconds_in_wait as blocked_duration_seconds,
			blocked_sql.sql_text as blocked_sql_text
		FROM v$session blocking
		JOIN v$session blocked ON blocking.sid = blocked.blocking_session
		LEFT JOIN v$sql blocking_sql ON blocking.sql_id = blocking_sql.sql_id
		LEFT JOIN v$sql blocked_sql ON blocked.sql_id = blocked_sql.sql_id
		WHERE blocking.type = 'USER'
		ORDER BY blocked.seconds_in_wait DESC
	`

	// QueryTablespaces retrieves tablespace usage information
	QueryTablespaces = `
		SELECT
			df.tablespace_name,
			df.total_size_mb,
			df.total_size_mb - NVL(fs.free_size_mb, 0) as used_size_mb,
			NVL(fs.free_size_mb, 0) as free_size_mb,
			ROUND(((df.total_size_mb - NVL(fs.free_size_mb, 0)) / df.total_size_mb) * 100, 2) as usage_percentage,
			ts.status,
			ts.contents,
			df.datafile_count
		FROM (
			SELECT 
				tablespace_name,
				ROUND(SUM(bytes) / 1024 / 1024, 2) as total_size_mb,
				COUNT(*) as datafile_count
			FROM dba_data_files
			GROUP BY tablespace_name
		) df
		LEFT JOIN (
			SELECT 
				tablespace_name,
				ROUND(SUM(bytes) / 1024 / 1024, 2) as free_size_mb
			FROM dba_free_space
			GROUP BY tablespace_name
		) fs ON df.tablespace_name = fs.tablespace_name
		JOIN dba_tablespaces ts ON df.tablespace_name = ts.tablespace_name
		ORDER BY usage_percentage DESC
	`

	// QueryTopSQLByElapsedTime retrieves top SQL by elapsed time
	QueryTopSQLByElapsedTime = `
		SELECT
			sql_id,
			SUBSTR(sql_text, 1, 4000) as sql_text,
			parsing_schema_name,
			executions,
			ROUND(elapsed_time / 1000000, 2) as elapsed_time_seconds,
			ROUND(elapsed_time / executions / 1000000, 4) as avg_elapsed_seconds,
			ROUND(cpu_time / 1000000, 2) as cpu_time_seconds,
			disk_reads,
			buffer_gets,
			rows_processed,
			first_load_time,
			last_active_time
		FROM v$sql
		WHERE executions > 0
		  AND parsing_schema_name IS NOT NULL
		ORDER BY elapsed_time DESC
		FETCH FIRST :1 ROWS ONLY
	`

	// QueryTopSQLByCPU retrieves top SQL by CPU time
	QueryTopSQLByCPU = `
		SELECT
			sql_id,
			SUBSTR(sql_text, 1, 4000) as sql_text,
			parsing_schema_name,
			executions,
			ROUND(cpu_time / 1000000, 2) as cpu_time_seconds,
			ROUND(cpu_time / executions / 1000000, 4) as avg_cpu_seconds,
			ROUND(elapsed_time / 1000000, 2) as elapsed_time_seconds,
			disk_reads,
			buffer_gets,
			rows_processed
		FROM v$sql
		WHERE executions > 0
		  AND parsing_schema_name IS NOT NULL
		ORDER BY cpu_time DESC
		FETCH FIRST :1 ROWS ONLY
	`

	// QueryDatabaseInstance retrieves database instance information
	QueryDatabaseInstance = `
		SELECT
			instance_name,
			host_name,
			version,
			startup_time,
			status,
			database_status,
			instance_role,
			ROUND((SYSDATE - startup_time), 2) as uptime_days
		FROM v$instance
	`

	// QueryDatabaseSize retrieves total database size
	QueryDatabaseSize = `
		SELECT
			ROUND(SUM(bytes) / 1024 / 1024 / 1024, 2) as total_size_gb
		FROM dba_data_files
	`

	// QuerySchemas retrieves all non-system schemas with object counts
	QuerySchemas = `
		SELECT
			owner as schema_name,
			COUNT(*) as total_objects,
			SUM(CASE WHEN object_type = 'TABLE' THEN 1 ELSE 0 END) as table_count,
			SUM(CASE WHEN object_type = 'INDEX' THEN 1 ELSE 0 END) as index_count,
			SUM(CASE WHEN object_type = 'VIEW' THEN 1 ELSE 0 END) as view_count,
			SUM(CASE WHEN object_type = 'PROCEDURE' THEN 1 ELSE 0 END) as procedure_count,
			SUM(CASE WHEN object_type = 'FUNCTION' THEN 1 ELSE 0 END) as function_count,
			SUM(CASE WHEN object_type = 'PACKAGE' THEN 1 ELSE 0 END) as package_count
		FROM dba_objects
		WHERE owner NOT IN ('SYS', 'SYSTEM', 'OUTLN', 'DBSNMP', 'WMSYS', 'XDB', 'CTXSYS', 'MDSYS', 'ORDSYS')
		GROUP BY owner
		ORDER BY total_objects DESC
	`

	// QueryInvalidObjects retrieves invalid objects
	QueryInvalidObjects = `
		SELECT
			owner as schema_name,
			object_name,
			object_type,
			status,
			last_ddl_time,
			created
		FROM dba_objects
		WHERE status = 'INVALID'
		  AND owner NOT IN ('SYS', 'SYSTEM', 'OUTLN', 'DBSNMP', 'WMSYS', 'XDB', 'CTXSYS', 'MDSYS', 'ORDSYS')
		ORDER BY owner, object_type, object_name
	`
)