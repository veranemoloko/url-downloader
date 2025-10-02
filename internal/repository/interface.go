package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/veranemoloko/url-downloader/internal/domain"
)

// TaskRepo defines the interface for task storage operations.
type TaskRepo interface {
	CreateTask(ctx context.Context, task *domain.Task) error
	GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	UpdateTask(ctx context.Context, task *domain.Task) error
	GetTasksByStatus(ctx context.Context, status domain.TaskStatus) ([]*domain.Task, error)
}
