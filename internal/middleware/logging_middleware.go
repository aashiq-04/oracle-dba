package middleware

import (
	"net/http"
	"time"

	"github.com/aashiq-04/oracle-dba/pkg/logger"
)

// LoggingMiddleware logs HTTP requests
type LoggingMiddleware struct {
	logger logger.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(log logger.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: log,
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Middleware returns an HTTP middleware function
func (m *LoggingMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log request
		duration := time.Since(start)
		m.logger.Info("HTTP request",
			logger.String("method", r.Method),
			logger.String("path", r.URL.Path),
			logger.Int("status", wrapped.statusCode),
			logger.Duration("duration", duration),
			logger.String("remote_addr", r.RemoteAddr),
		)
	})
}