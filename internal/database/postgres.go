package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresDB wraps the PostgreSQL connection pool
type PostgresDB struct {
	DB *sql.DB
}

// PostgresConfig holds PostgreSQL connection configuration
type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int
	MinConns int
}

// NewPostgresDB creates a new PostgreSQL connection pool
func NewPostgresDB(cfg PostgresConfig) (*PostgresDB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.MaxConns)
	db.SetMaxIdleConns(cfg.MinConns)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(10 * time.Minute)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &PostgresDB{DB: db}, nil
}

// Close closes the PostgreSQL connection pool
func (p *PostgresDB) Close() error {
	return p.DB.Close()
}

// Health checks the health of the PostgreSQL connection
func (p *PostgresDB) Health(ctx context.Context) error {
	return p.DB.PingContext(ctx)
}

// Stats returns connection pool statistics
func (p *PostgresDB) Stats() sql.DBStats {
	return p.DB.Stats()
}