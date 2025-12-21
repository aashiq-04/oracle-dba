package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type sessionMetricsRepository struct {
	db *sql.DB
}

// NewSessionMetricsRepository creates a new session metrics repository
func NewSessionMetricsRepository(db *sql.DB) SessionMetricsRepository {
	return &sessionMetricsRepository{db: db}
}

func (r *sessionMetricsRepository) Create(ctx context.Context, metrics []*SessionMetric) error {
	if len(metrics) == 0 {
		return nil
	}

	query := `
		INSERT INTO monitoring.session_snapshots (
			id, oracle_db, snapshot_time, active_sessions, blocked_sessions, raw_data
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, metric := range metrics {
		metric.ID = uuid.New()
		metric.CapturedAt = time.Now()

		// For now, store basic counts in raw_data as JSON
		rawData := fmt.Sprintf(`{"sid": %d, "serial": %d, "status": "%s"}`,
			metric.OracleSID, metric.OracleSerial, metric.Status)

		// Determine if session is active/blocked
		activeCount := 0
		if metric.Status == "ACTIVE" {
			activeCount = 1
		}
		blockedCount := 0
		if metric.BlockingSession != nil {
			blockedCount = 1
		}

		_, err := tx.ExecContext(ctx, query,
			metric.ID,
			"default", // Oracle DB name
			metric.CapturedAt,
			activeCount,
			blockedCount,
			rawData,
		)
		if err != nil {
			return fmt.Errorf("failed to insert session metric: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *sessionMetricsRepository) GetByTimeRange(ctx context.Context, start, end time.Time) ([]*SessionMetric, error) {
	query := `
		SELECT id, oracle_db, snapshot_time, active_sessions, blocked_sessions
		FROM monitoring.session_snapshots
		WHERE snapshot_time BETWEEN $1 AND $2
		ORDER BY snapshot_time DESC
	`

	rows, err := r.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query session metrics: %w", err)
	}
	defer rows.Close()

	metrics := []*SessionMetric{}
	for rows.Next() {
		metric := &SessionMetric{}
		var oracleDB string
		var activeSessions, blockedSessions int

		err := rows.Scan(
			&metric.ID,
			&oracleDB,
			&metric.CapturedAt,
			&activeSessions,
			&blockedSessions,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session metric: %w", err)
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (r *sessionMetricsRepository) GetBySchema(ctx context.Context, schema string, start, end time.Time) ([]*SessionMetric, error) {
	// Simplified implementation - would need to parse raw_data in production
	return r.GetByTimeRange(ctx, start, end)
}