package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddress string
	DownloadDir   string
	TaskDir       string
	MaxWorkers    int
	SaveInterval  time.Duration
	LogLevel      string
}

// Load reads environment variables (optionally from a .env file) and
// returns a Config struct with default values applied if variables are missing.
// It also ensures that the download and task directories exist.
func Load() (*Config, error) {

	if err := godotenv.Load(); err != nil {
		fmt.Println("Note: .env file not found, using defaults")
	}

	cfg := &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", ":8080"),
		DownloadDir:   getEnv("DOWNLOAD_DIR", "downloads/files"),
		TaskDir:       getEnv("TASK_DIR", "downloads/tasks"),
		MaxWorkers:    getEnvAsInt("MAX_WORKERS", 5),
		SaveInterval:  getEnvAsDuration("SAVE_INTERVAL", time.Second*10),
		LogLevel:      getEnv("LOG_LEVEL", "INFO"),
	}

	if err := os.MkdirAll(cfg.DownloadDir, 0755); err != nil {
		return nil, fmt.Errorf("create download dir: %w", err)
	}
	if err := os.MkdirAll(cfg.TaskDir, 0755); err != nil {
		return nil, fmt.Errorf("create task dir: %w", err)
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
