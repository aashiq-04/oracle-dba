package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	_ = godotenv.Load()
	// Database connection string
	dbHost := getEnv("POSTGRES_HOST", "localhost")
	dbPort := getEnv("POSTGRES_PORT", "5432")
	dbUser := getEnv("POSTGRES_USER", "postgres")
	dbPassword := getEnv("POSTGRES_PASSWORD", "")
	dbName := getEnv("POSTGRES_DB", "oracle_dba_platform")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	// Connect to database
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Get admin credentials from arguments or use defaults
	username := "admin"
	email := "admin@oracleplatform.com"
	password := "admin123"

	if len(os.Args) > 1 {
		username = os.Args[1]
	}
	if len(os.Args) > 2 {
		email = os.Args[2]
	}
	if len(os.Args) > 3 {
		password = os.Args[3]
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create user
	var userID string
	err = db.QueryRow(`
		INSERT INTO auth.users (username, email, password_hash, is_active)
		VALUES ($1, $2, $3, true)
		ON CONFLICT (username) DO UPDATE
		SET password_hash = EXCLUDED.password_hash,
		    email = EXCLUDED.email
		RETURNING id
	`, username, email, string(hashedPassword)).Scan(&userID)

	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("✓ User created/updated: %s (ID: %s)\n", username, userID)

	// Get ADMIN role ID
	var adminRoleID string
	err = db.QueryRow(`SELECT id FROM auth.roles WHERE name = 'ADMIN'`).Scan(&adminRoleID)
	if err != nil {
		log.Fatalf("Failed to get ADMIN role: %v", err)
	}

	// Assign ADMIN role
	_, err = db.Exec(`
		INSERT INTO auth.user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, adminRoleID)

	if err != nil {
		log.Fatalf("Failed to assign ADMIN role: %v", err)
	}

	fmt.Println("✓ ADMIN role assigned")
	fmt.Println()
	fmt.Println("=== Admin User Created Successfully ===")
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Email:    %s\n", email)
	fmt.Printf("Password: %s\n", password)
	fmt.Println()
	fmt.Println("You can now login using these credentials.")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}