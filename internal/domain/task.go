package domain

import (
	"time"

	"github.com/google/uuid"
)

// Task represents a download task containing multiple URLs and their statuses.
type Task struct {
	ID        uuid.UUID      `json:"task_id"`
	Status    TaskStatus     `json:"status"`
	URLs      []string       `json:"urls"`
	Downloads []DownloadItem `json:"downloads"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// DownloadItem represents a single URL download and its current status.
type DownloadItem struct {
	URL      string         `json:"url"`
	Status   DownloadStatus `json:"status"`
	FilePath string         `json:"file_path,omitempty"`
	Error    string         `json:"error,omitempty"`
}
