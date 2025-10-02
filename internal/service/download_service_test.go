package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/veranemoloko/url-downloader/internal/config"
	"github.com/veranemoloko/url-downloader/internal/domain"
)

type mockRepoDS struct{}

func (m *mockRepoDS) CreateTask(ctx context.Context, task *domain.Task) error { return nil }
func (m *mockRepoDS) GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	return nil, nil
}
func (m *mockRepoDS) UpdateTask(ctx context.Context, task *domain.Task) error { return nil }
func (m *mockRepoDS) GetTasksByStatus(ctx context.Context, status domain.TaskStatus) ([]*domain.Task, error) {
	return nil, nil
}

func TestDownloadService_ProcessTask(t *testing.T) {
	task := &domain.Task{
		ID:     uuid.New(),
		Status: domain.TaskStatusPending,
		URLs:   []string{"http://a"},
		Downloads: []domain.DownloadItem{
			{URL: "http://a", Status: domain.DownloadStatusPending},
		},
	}

	cfg := &config.Config{
		WorkerPoolSize:  1,
		DownloadTimeout: time.Second,
		DownloadDir:     t.TempDir(),
		MaxFileSize:     1024,
	}

	ds := NewDownloadService(&mockRepoDS{}, cfg)
	err := ds.ProcessTask(context.Background(), task)

	assert.NoError(t, err)
	assert.Equal(t, domain.TaskStatusInProgress, task.Status)
}

func TestDownloadService_isTaskCompleted(t *testing.T) {
	task := &domain.Task{
		Downloads: []domain.DownloadItem{
			{Status: domain.DownloadStatusCompleted},
		},
	}
	ds := NewDownloadService(&mockRepoDS{}, &config.Config{WorkerPoolSize: 1, DownloadTimeout: time.Second})

	assert.True(t, ds.isTaskCompleted(task))
}
