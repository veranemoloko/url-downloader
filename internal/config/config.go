package config

import (
	"fmt"
	"time"
)

// Config holds all application configuration settings.
type Config struct {
	Environment string `envconfig:"FD_ENV" default:"development"`

	HTTPPort    int           `envconfig:"FD_HTTP_PORT" default:"8080"`
	HTTPTimeout time.Duration `envconfig:"FD_HTTP_TIMEOUT" default:"15s"`

	WorkerPoolSize  int           `envconfig:"FD_WORKER_POOL_SIZE" default:"5"`
	MaxURLsPerTask  int           `envconfig:"FD_MAX_URLS_PER_TASK" default:"10"`
	DownloadTimeout time.Duration `envconfig:"FD_DOWNLOAD_TIMEOUT" default:"5m"`
	MaxFileSize     int64         `envconfig:"FD_MAX_FILE_SIZE" default:"104857600"`

	DownloadDir string `envconfig:"FD_DOWNLOAD_DIR" default:"./storage"`
	StateFile   string `envconfig:"FD_STATE_FILE" default:"./state.json"`
	TempDir     string `envconfig:"FD_TEMP_DIR" default:"./tmp"`

	ShutdownTimeout time.Duration `envconfig:"FD_SHUTDOWN_TIMEOUT" default:"30s"`

	LogLevel  string `envconfig:"FD_LOG_LEVEL" default:"info"`
	LogFormat string `envconfig:"FD_LOG_FORMAT" default:"json"`
}

// Validate checks the configuration for invalid or missing values.
// Returns an error describing the first invalid setting found.
func (c *Config) Validate() error {
	if c.HTTPPort <= 0 || c.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", c.HTTPPort)
	}

	if c.WorkerPoolSize <= 0 {
		return fmt.Errorf("worker pool size must be positive: %d", c.WorkerPoolSize)
	}

	if c.MaxURLsPerTask <= 0 {
		return fmt.Errorf("max URLs per task must be positive: %d", c.MaxURLsPerTask)
	}

	if c.MaxFileSize <= 0 {
		return fmt.Errorf("max file size must be positive: %d", c.MaxFileSize)
	}

	if c.DownloadDir == "" {
		return fmt.Errorf("download directory cannot be empty")
	}
	if c.StateFile == "" {
		return fmt.Errorf("state file cannot be empty")
	}
	if c.TempDir == "" {
		return fmt.Errorf("temp directory cannot be empty")
	}

	return nil
}
