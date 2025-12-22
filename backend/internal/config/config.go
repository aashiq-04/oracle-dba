package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server    ServerConfig
	Postgres  PostgresConfig
	Oracle    OracleConfig
	JWT       JWTConfig
	Logging   LoggingConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port            string
	Host            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
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

// JWTConfig holds JWT token configuration
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
	Issuer     string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, text
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file (optional, ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", "8080"),
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		Postgres: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", ""),
			DBName:   getEnv("POSTGRES_DB", "oracle_dba_platform"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
			MaxConns: getIntEnv("POSTGRES_MAX_CONNS", 25),
			MinConns: getIntEnv("POSTGRES_MIN_CONNS", 5),
		},
		Oracle: OracleConfig{
			Host:        getEnv("ORACLE_HOST", "localhost"),
			Port:        getEnv("ORACLE_PORT", "1521"),
			ServiceName: getEnv("ORACLE_SERVICE_NAME", "ORCLPDB1"),
			Username:    getEnv("ORACLE_USERNAME", ""),
			Password:    getEnv("ORACLE_PASSWORD", ""),
			MaxConns:    getIntEnv("ORACLE_MAX_CONNS", 10),
			MinConns:    getIntEnv("ORACLE_MIN_CONNS", 2),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", ""),
			Expiration: getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
			Issuer:     getEnv("JWT_ISSUER", "oracle-dba-platform"),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	// Validate critical configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate PostgreSQL
	if c.Postgres.Password == "" {
		return fmt.Errorf("POSTGRES_PASSWORD is required")
	}

	// Validate Oracle
	if c.Oracle.Username == "" {
		return fmt.Errorf("ORACLE_USERNAME is required")
	}
	if c.Oracle.Password == "" {
		return fmt.Errorf("ORACLE_PASSWORD is required")
	}

	// Validate JWT
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	return nil
}

// PostgresDSN returns PostgreSQL connection string
func (c *PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// OracleDSN returns Oracle connection string
func (c *OracleConfig) DSN() string {
	return fmt.Sprintf(
		"%s/%s@%s:%s/%s",
		c.Username, c.Password, c.Host, c.Port, c.ServiceName,
	)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}