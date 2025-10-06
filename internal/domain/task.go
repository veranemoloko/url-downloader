package domain

import (
	"time"
)

// TaskStatus represents the current state of a download task.
type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusInProgress TaskStatus = "inprogress"
	StatusCompleted  TaskStatus = "completed"
	StatusFailed     TaskStatus = "failed"
)

// Task represents a download task containing multiple URLs and their results.
type Task struct {
	ID        string           `json:"id"`
	URLs      []string         `json:"urls"`
	Status    TaskStatus       `json:"status"`
	Results   []DownloadResult `json:"results,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// DownloadResult represents the outcome of downloading a single URL.
type DownloadResult struct {
	URL       string `json:"url"`
	FileName  string `json:"file_name,omitempty"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	BytesRead int64  `json:"bytes_read"`
	Hash      string `json:"hash,omitempty"`
}

// TaskEvent represents an event related to a task, used for notifications or updates.
type TaskEvent struct {
	Type    EventType
	TaskID  string
	Task    *Task
	Updates *TaskUpdate
}

// EventType represents the type of task event.
type EventType string

const (
	EventCreateTask EventType = "create"
	EventUpdateTask EventType = "update"
)

// TaskUpdate represents updates applied to a task, such as status changes or download results.
type TaskUpdate struct {
	Status  *TaskStatus
	Results []DownloadResult
}
