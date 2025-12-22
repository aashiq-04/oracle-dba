package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/aashiq-04/oracle-dba/internal/config"
	"github.com/aashiq-04/oracle-dba/internal/database"
	"github.com/aashiq-04/oracle-dba/internal/graph"
	"github.com/aashiq-04/oracle-dba/internal/middleware"
	"github.com/aashiq-04/oracle-dba/internal/repository"
	"github.com/aashiq-04/oracle-dba/internal/service"
	"github.com/aashiq-04/oracle-dba/pkg/logger"
	"github.com/aashiq-04/oracle-dba/pkg/oracle"
)

func main() {
	// Initialize logger
	log := logger.NewLogger()
	log.Info("Starting Oracle DBA Platform...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", logger.Error(err))
	}
	log.Info("Configuration loaded successfully")

	// Connect to PostgreSQL
	log.Info("Connecting to PostgreSQL...")
	pgDB, err := database.NewPostgresDB(database.PostgresConfig{
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		User:     cfg.Postgres.User,
		Password: cfg.Postgres.Password,
		DBName:   cfg.Postgres.DBName,
		SSLMode:  cfg.Postgres.SSLMode,
		MaxConns: cfg.Postgres.MaxConns,
		MinConns: cfg.Postgres.MinConns,
	})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL", logger.Error(err))
	}
	defer pgDB.Close()
	log.Info("PostgreSQL connected successfully")

	// Connect to Oracle
	log.Info("Connecting to Oracle Database...")
	oracleDB, err := oracle.NewOracleDB(oracle.OracleConfig{
		Host:        cfg.Oracle.Host,
		Port:        cfg.Oracle.Port,
		ServiceName: cfg.Oracle.ServiceName,
		Username:    cfg.Oracle.Username,
		Password:    cfg.Oracle.Password,
		MaxConns:    cfg.Oracle.MaxConns,
		MinConns:    cfg.Oracle.MinConns,
	})
	if err != nil {
		log.Fatal("Failed to connect to Oracle", logger.Error(err))
	}
	defer oracleDB.Close()
	log.Info("Oracle Database connected successfully")

	// Initialize repositories
	log.Info("Initializing repositories...")
	repos := &repository.Repositories{
		Users:             repository.NewUserRepository(pgDB.DB),
		Roles:             repository.NewRoleRepository(pgDB.DB),
		UserRoles:         repository.NewUserRoleRepository(pgDB.DB),
		Permissions:       repository.NewPermissionRepository(pgDB.DB),
		AuditLogs:         repository.NewAuditLogRepository(pgDB.DB),
		SessionMetrics:    repository.NewSessionMetricsRepository(pgDB.DB),
		TablespaceMetrics: repository.NewTablespaceMetricsRepository(pgDB.DB),
		QueryMetrics:      repository.NewQueryMetricsRepository(pgDB.DB),
	}
	log.Info("Repositories initialized successfully")

	// Initialize services
	log.Info("Initializing services...")
	authService := service.NewAuthService(
		repos.Users,
		repos.UserRoles,
		repos.Permissions,
		repos.AuditLogs,
		cfg.JWT.Secret,
		cfg.JWT.Expiration,
		cfg.JWT.Issuer,
	)

	rbacService := service.NewRBACService(
		repos.UserRoles,
		repos.Permissions,
		repos.Roles,
		repos.AuditLogs,
	)

	oracleService := service.NewOracleService(
		oracleDB,
		repos.SessionMetrics,
		repos.TablespaceMetrics,
		repos.QueryMetrics,
		repos.AuditLogs,
	)
	log.Info("Services initialized successfully")

	// Initialize GraphQL resolver
	resolver := graph.NewResolver(authService, rbacService, oracleService)

	// Create GraphQL server
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
		Resolvers: resolver,
	}))

	// Setup middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)
	corsMiddleware := middleware.NewCORSMiddleware()
	loggingMiddleware := middleware.NewLoggingMiddleware(log)

	// Create HTTP router
	mux := http.NewServeMux()

	// GraphQL endpoint with middleware chain
	mux.Handle("/query",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				authMiddleware.Middleware(srv),
			),
		),
	)

	// GraphQL Playground (development only)
	mux.Handle("/", playground.Handler("Oracle DBA Platform", "/query"))

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Check PostgreSQL health
		if err := pgDB.Health(r.Context()); err != nil {
			http.Error(w, "PostgreSQL unhealthy", http.StatusServiceUnavailable)
			return
		}

		// Check Oracle health
		if err := oracleDB.Health(r.Context()); err != nil {
			http.Error(w, "Oracle unhealthy", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create HTTP server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Info(fmt.Sprintf("Server starting on http://%s", addr))
		log.Info(fmt.Sprintf("GraphQL Playground available at http://%s/", addr))
		log.Info(fmt.Sprintf("GraphQL API endpoint: http://%s/query", addr))

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", logger.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", logger.Error(err))
	}

	log.Info("Server stopped")
}