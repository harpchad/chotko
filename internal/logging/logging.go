// Package logging provides structured logging for the chotko application.
// It uses Go's log/slog package for structured, leveled logging.
package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// Logger is the global application logger.
var Logger *slog.Logger

// Config holds logging configuration.
type Config struct {
	Level   slog.Level
	Output  io.Writer
	JSON    bool
	LogFile string
}

// DefaultConfig returns the default logging configuration.
// Logs to stderr at Info level in text format.
func DefaultConfig() Config {
	return Config{
		Level:  slog.LevelInfo,
		Output: os.Stderr,
		JSON:   false,
	}
}

// Init initializes the global logger with the given configuration.
func Init(cfg Config) error {
	var handler slog.Handler
	output := cfg.Output

	// If log file specified, open it
	if cfg.LogFile != "" {
		dir := filepath.Dir(cfg.LogFile)
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return err
		}
		f, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			return err
		}
		output = f
	}

	opts := &slog.HandlerOptions{
		Level: cfg.Level,
	}

	if cfg.JSON {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	Logger = slog.New(handler)
	return nil
}

// Debug logs a debug message.
func Debug(msg string, args ...any) {
	if Logger != nil {
		Logger.Debug(msg, args...)
	}
}

// Info logs an info message.
func Info(msg string, args ...any) {
	if Logger != nil {
		Logger.Info(msg, args...)
	}
}

// Warn logs a warning message.
func Warn(msg string, args ...any) {
	if Logger != nil {
		Logger.Warn(msg, args...)
	}
}

// Error logs an error message.
func Error(msg string, args ...any) {
	if Logger != nil {
		Logger.Error(msg, args...)
	}
}

// With returns a logger with additional context.
func With(args ...any) *slog.Logger {
	if Logger != nil {
		return Logger.With(args...)
	}
	return slog.Default()
}
