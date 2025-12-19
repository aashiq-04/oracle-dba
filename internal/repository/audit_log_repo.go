package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type auditLogRepository struct {
	db *sql.DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *sql.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, log *AuditLog) error {
	query := `
		INSERT INTO audit.logs (
			id, user_id, username, action, target, success, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	log.ID = uuid.New()
	log.Timestamp = time.Now()

	// Convert status to boolean for success field
	success := log.Status == "SUCCESS"

	// Build metadata JSON (simplified - you can enhance this)
	metadata := fmt.Sprintf(`{"resource_type": "%s", "resource_id": "%s", "oracle_schema": "%s"}`,
		log.ResourceType,
		safeString(log.ResourceID),
		safeString(log.OracleSchema),
	)

	_, err := r.db.ExecContext(ctx, query,
		log.ID,
		log.UserID,
		log.Username,
		log.Action,
		log.ResourceType, // Using ResourceType as target
		success,
		metadata,
		log.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

func (r *auditLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*AuditLog, error) {
	query := `
		SELECT id, user_id, username, action, target, success, metadata, created_at
		FROM audit.logs
		WHERE id = $1
	`

	log := &AuditLog{}
	var success bool
	var metadata string
	var target string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID,
		&log.UserID,
		&log.Username,
		&log.Action,
		&target,
		&success,
		&metadata,
		&log.Timestamp,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("audit log not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	// Convert boolean to status string
	if success {
		log.Status = "SUCCESS"
	} else {
		log.Status = "FAILURE"
	}

	log.ResourceType = target

	return log, nil
}

func (r *auditLogRepository) List(ctx context.Context, filter *AuditLogFilter) ([]*AuditLog, error) {
	query := `
		SELECT id, user_id, username, action, target, success, metadata, created_at
		FROM audit.logs
		WHERE 1=1
	`

	args := []interface{}{}
	argCounter := 1

	// Build dynamic WHERE clause
	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argCounter)
		args = append(args, *filter.UserID)
		argCounter++
	}

	if filter.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", argCounter)
		args = append(args, *filter.Action)
		argCounter++
	}

	if filter.ResourceType != nil {
		query += fmt.Sprintf(" AND target = $%d", argCounter)
		args = append(args, *filter.ResourceType)
		argCounter++
	}

	if filter.StartTime != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argCounter)
		args = append(args, *filter.StartTime)
		argCounter++
	}

	if filter.EndTime != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argCounter)
		args = append(args, *filter.EndTime)
		argCounter++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCounter)
		args = append(args, filter.Limit)
		argCounter++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCounter)
		args = append(args, filter.Offset)
		argCounter++
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer rows.Close()

	logs := []*AuditLog{}
	for rows.Next() {
		log := &AuditLog{}
		var success bool
		var metadata string
		var target string

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Username,
			&log.Action,
			&target,
			&success,
			&metadata,
			&log.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		// Convert boolean to status string
		if success {
			log.Status = "SUCCESS"
		} else {
			log.Status = "FAILURE"
		}

		log.ResourceType = target
		logs = append(logs, log)
	}

	return logs, nil
}

func (r *auditLogRepository) Count(ctx context.Context, filter *AuditLogFilter) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM audit.logs
		WHERE 1=1
	`

	args := []interface{}{}
	argCounter := 1

	// Build dynamic WHERE clause (same as List)
	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argCounter)
		args = append(args, *filter.UserID)
		argCounter++
	}

	if filter.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", argCounter)
		args = append(args, *filter.Action)
		argCounter++
	}

	if filter.ResourceType != nil {
		query += fmt.Sprintf(" AND target = $%d", argCounter)
		args = append(args, *filter.ResourceType)
		argCounter++
	}

	if filter.StartTime != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argCounter)
		args = append(args, *filter.StartTime)
		argCounter++
	}

	if filter.EndTime != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argCounter)
		args = append(args, *filter.EndTime)
		argCounter++
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return count, nil
}

// Helper function to safely dereference string pointers
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return strings.ReplaceAll(*s, `"`, `\"`)
}