package domain

import (
	"time"

	"github.com/google/uuid"
)

// CreateTaskRequest represents the request body for creating a new Task.
type CreateTaskRequest struct {
	URLs []string `json:"urls" validate:"required,min=1,max=10,dive,url"`
}

// TaskResponse represents the response returned for a Task, including its status and downloads.
type TaskResponse struct {
	ID        uuid.UUID      `json:"task_id"`
	Status    TaskStatus     `json:"status"`
	Downloads []DownloadItem `json:"downloads"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}
