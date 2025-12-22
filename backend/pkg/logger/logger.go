package logger

import (
	"context"
	"log"
	"os"
	"time"
)

// Logger interface defines logging methods
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
}

// Field represents a structured logging field
type Field struct {
	Key   string
	Value interface{}
}

// String creates a string field
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an integer field
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Error creates an error field
func Error(err error) Field {
	return Field{Key: "error", Value: err.Error()}
}

// Duration creates a duration field
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value.String()}
}

// SimpleLogger is a basic logger implementation
type SimpleLogger struct {
	logger *log.Logger
	fields []Field
}

// NewLogger creates a new logger instance
func NewLogger() Logger {
	return &SimpleLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
		fields: []Field{},
	}
}

func (l *SimpleLogger) Debug(msg string, fields ...Field) {
	l.log("DEBUG", msg, fields...)
}

func (l *SimpleLogger) Info(msg string, fields ...Field) {
	l.log("INFO", msg, fields...)
}

func (l *SimpleLogger) Warn(msg string, fields ...Field) {
	l.log("WARN", msg, fields...)
}

func (l *SimpleLogger) Error(msg string, fields ...Field) {
	l.log("ERROR", msg, fields...)
}

func (l *SimpleLogger) Fatal(msg string, fields ...Field) {
	l.log("FATAL", msg, fields...)
	os.Exit(1)
}

func (l *SimpleLogger) With(fields ...Field) Logger {
	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)
	return &SimpleLogger{
		logger: l.logger,
		fields: newFields,
	}
}

func (l *SimpleLogger) log(level, msg string, fields ...Field) {
	allFields := append(l.fields, fields...)
	fieldStr := ""
	for _, f := range allFields {
		fieldStr += " " + f.Key + "=" + formatValue(f.Value)
	}
	l.logger.Printf("[%s] %s%s", level, msg, fieldStr)
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int, int64:
		return string(rune(val.(int)))
	default:
		return ""
	}
}

// contextKey is a custom type for context keys
type contextKey string

const loggerKey contextKey = "logger"

// WithContext adds logger to context
func WithContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext retrieves logger from context
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		return logger
	}
	return NewLogger()
}