package domain

import (
	"time"
)

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusInProgress TaskStatus = "inprogress"
	StatusCompleted  TaskStatus = "completed"
	StatusFailed     TaskStatus = "failed"
)

type Task struct {
	ID        string           `json:"id"`
	URLs      []string         `json:"urls"`
	Status    TaskStatus       `json:"status"`
	Results   []DownloadResult `json:"results,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

type DownloadResult struct {
	URL       string `json:"url"`
	FileName  string `json:"file_name,omitempty"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	BytesRead int64  `json:"bytes_read"`
}

type TaskEvent struct {
	Type    EventType
	TaskID  string
	Task    *Task
	Updates *TaskUpdate
}

type EventType string

const (
	EventCreateTask EventType = "create"
	EventUpdateTask EventType = "update"
)

type TaskUpdate struct {
	Status  *TaskStatus
	Results []DownloadResult
}
