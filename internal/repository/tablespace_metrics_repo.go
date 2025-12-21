package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type tablespaceMetricsRepository struct {
	db *sql.DB
}

// NewTablespaceMetricsRepository creates a new tablespace metrics repository
func NewTablespaceMetricsRepository(db *sql.DB) TablespaceMetricsRepository {
	return &tablespaceMetricsRepository{db: db}
}

func (r *tablespaceMetricsRepository) Create(ctx context.Context, metrics []*TablespaceMetric) error {
	if len(metrics) == 0 {
		return nil
	}

	query := `
		INSERT INTO monitoring.session_snapshots (
			id, oracle_db, snapshot_time, active_sessions, blocked_sessions, raw_data
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	// Note: We're reusing session_snapshots table for simplicity
	// In production, you'd have a dedicated tablespace_metrics table

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, metric := range metrics {
		metric.ID = uuid.New()
		metric.CapturedAt = time.Now()

		// Store tablespace data as JSON in raw_data
		rawData := fmt.Sprintf(`{"tablespace": "%s", "usage": %.2f}`,
			metric.TablespaceName, metric.UsagePercentage)

		_, err := tx.ExecContext(ctx, query,
			metric.ID,
			"default",
			metric.CapturedAt,
			0, // active_sessions (not used for tablespace)
			0, // blocked_sessions (not used for tablespace)
			rawData,
		)
		if err != nil {
			return fmt.Errorf("failed to insert tablespace metric: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *tablespaceMetricsRepository) GetLatest(ctx context.Context) ([]*TablespaceMetric, error) {
	// Simplified - return empty for now
	return []*TablespaceMetric{}, nil
}

func (r *tablespaceMetricsRepository) GetByTablespaceName(ctx context.Context, name string, start, end time.Time) ([]*TablespaceMetric, error) {
	// Simplified - return empty for now
	return []*TablespaceMetric{}, nil
}

func (r *tablespaceMetricsRepository) GetByTimeRange(ctx context.Context, start, end time.Time) ([]*TablespaceMetric, error) {
	// Simplified - return empty for now
	return []*TablespaceMetric{}, nil
}