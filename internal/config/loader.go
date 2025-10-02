package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
)

// Load loads configuration from environment variables, validates it, and ensures required directories exist.
// Returns a pointer to Config or an error if loading or validation fails.
func Load() (*Config, error) {
	var cfg Config

	if err := envconfig.Process("UD", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process environment variables: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	if err := createDirs(&cfg); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	return &cfg, nil
}

func createDirs(cfg *Config) error {
	dirs := []string{
		cfg.DownloadDir,
		cfg.TempDir,
		filepath.Dir(cfg.StateFile),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		slog.Debug("directory created or verified", "path", dir)
	}
	return nil
}

// SetupLogger configures the global slog logger based on configuration.
// Supports "json" or "text" formats and log levels: debug, info, warn, error.
func SetupLogger(cfg *Config) {
	var level slog.Level
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if cfg.LogFormat == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
