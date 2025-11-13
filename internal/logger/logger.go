package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

// Initialize sets up the global logger with configuration from environment
func Initialize() {
	// Set up console writer for human-readable output in development
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	// Parse log level from environment, default to INFO
	logLevel := getLogLevel()
	zerolog.SetGlobalLevel(logLevel)

	// Create logger
	Logger = zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()

	log.Logger = Logger
}

// InitializeJSON sets up the logger for JSON output (production mode)
func InitializeJSON() {
	// Parse log level from environment, default to INFO
	logLevel := getLogLevel()
	zerolog.SetGlobalLevel(logLevel)

	// Create JSON logger
	Logger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Caller().
		Logger()

	log.Logger = Logger
}

// getLogLevel parses the LOG_LEVEL environment variable
func getLogLevel() zerolog.Level {
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		return zerolog.InfoLevel
	}

	switch strings.ToLower(levelStr) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	case "trace":
		return zerolog.TraceLevel
	default:
		return zerolog.InfoLevel
	}
}

// WithRequestID creates a logger with a request ID for tracing
func WithRequestID(requestID string) zerolog.Logger {
	return Logger.With().Str("request_id", requestID).Logger()
}

// Debug logs a debug message
func Debug(msg string) {
	Logger.Debug().Msg(msg)
}

// Info logs an info message
func Info(msg string) {
	Logger.Info().Msg(msg)
}

// Warn logs a warning message
func Warn(msg string) {
	Logger.Warn().Msg(msg)
}

// Error logs an error message
func Error(msg string) {
	Logger.Error().Msg(msg)
}

// Fatal logs a fatal message and exits
func Fatal(msg string) {
	Logger.Fatal().Msg(msg)
}

// GetLogger returns the global logger instance
func GetLogger() *zerolog.Logger {
	return &Logger
}
