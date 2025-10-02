package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/veranemoloko/url-downloader/internal/domain"
)

func TestTaskStorage_CRUD(t *testing.T) {
	file := t.TempDir() + "/tasks.json"
	repo, err := NewTaskStorage(file)
	assert.NoError(t, err)

	task := &domain.Task{
		ID:     uuid.New(),
		Status: domain.TaskStatusPending,
		URLs:   []string{"http://a"},
	}

	err = repo.CreateTask(context.Background(), task)
	assert.NoError(t, err)

	got, err := repo.GetTask(context.Background(), task.ID)
	assert.NoError(t, err)
	assert.Equal(t, task.ID, got.ID)

	task.Status = domain.TaskStatusCompleted
	err = repo.UpdateTask(context.Background(), task)
	assert.NoError(t, err)

	got2, err := repo.GetTask(context.Background(), task.ID)
	assert.NoError(t, err)
	assert.Equal(t, domain.TaskStatusCompleted, got2.Status)
}

func TestTaskStorage_GetTasksByStatus(t *testing.T) {
	file := t.TempDir() + "/tasks.json"
	repo, err := NewTaskStorage(file)
	assert.NoError(t, err)

	task1 := &domain.Task{ID: uuid.New(), Status: domain.TaskStatusPending}
	task2 := &domain.Task{ID: uuid.New(), Status: domain.TaskStatusCompleted}

	_ = repo.CreateTask(context.Background(), task1)
	_ = repo.CreateTask(context.Background(), task2)

	pending, err := repo.GetTasksByStatus(context.Background(), domain.TaskStatusPending)
	assert.NoError(t, err)
	assert.Len(t, pending, 1)
	assert.Equal(t, task1.ID, pending[0].ID)

	completed, err := repo.GetTasksByStatus(context.Background(), domain.TaskStatusCompleted)
	assert.NoError(t, err)
	assert.Len(t, completed, 1)
	assert.Equal(t, task2.ID, completed[0].ID)
}
