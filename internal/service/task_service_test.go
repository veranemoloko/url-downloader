package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/veranemoloko/url-downloader/internal/domain"
	"github.com/veranemoloko/url-downloader/internal/service"
)

type mockRepo struct {
	createTaskFunc     func(ctx context.Context, task *domain.Task) error
	getTaskFunc        func(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	updateTaskFunc     func(ctx context.Context, task *domain.Task) error
	getTasksByStatusFn func(ctx context.Context, status domain.TaskStatus) ([]*domain.Task, error)
}

func (m *mockRepo) CreateTask(ctx context.Context, task *domain.Task) error {
	return m.createTaskFunc(ctx, task)
}
func (m *mockRepo) GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	return m.getTaskFunc(ctx, id)
}
func (m *mockRepo) UpdateTask(ctx context.Context, task *domain.Task) error {
	return m.updateTaskFunc(ctx, task)
}
func (m *mockRepo) GetTasksByStatus(ctx context.Context, status domain.TaskStatus) ([]*domain.Task, error) {
	if m.getTasksByStatusFn != nil {
		return m.getTasksByStatusFn(ctx, status)
	}
	return nil, nil
}

func TestTaskService_CreateTask_Success(t *testing.T) {
	repoCalled := false
	mockRepo := &mockRepo{
		createTaskFunc: func(ctx context.Context, task *domain.Task) error {
			repoCalled = true
			return nil
		},
		updateTaskFunc: func(ctx context.Context, task *domain.Task) error { return nil },
		getTaskFunc:    func(ctx context.Context, id uuid.UUID) (*domain.Task, error) { return nil, nil },
	}

	svc := service.NewTaskService(mockRepo, nil)
	req := &domain.CreateTaskRequest{URLs: []string{"http://a", "http://b"}}

	task, err := svc.CreateTask(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.True(t, repoCalled)
}

func TestTaskService_CreateTask_RepoError(t *testing.T) {
	mockRepo := &mockRepo{
		createTaskFunc: func(ctx context.Context, task *domain.Task) error { return errors.New("db error") },
		updateTaskFunc: func(ctx context.Context, task *domain.Task) error { return nil },
		getTaskFunc:    func(ctx context.Context, id uuid.UUID) (*domain.Task, error) { return nil, nil },
	}

	svc := service.NewTaskService(mockRepo, nil)
	req := &domain.CreateTaskRequest{URLs: []string{"http://a"}}

	task, err := svc.CreateTask(context.Background(), req)
	assert.Nil(t, task)
	assert.Error(t, err)
}

func TestTaskService_GetTask(t *testing.T) {
	expected := &domain.Task{ID: uuid.New()}
	mockRepo := &mockRepo{
		getTaskFunc:    func(ctx context.Context, id uuid.UUID) (*domain.Task, error) { return expected, nil },
		createTaskFunc: func(ctx context.Context, task *domain.Task) error { return nil },
		updateTaskFunc: func(ctx context.Context, task *domain.Task) error { return nil },
	}

	svc := service.NewTaskService(mockRepo, nil)
	task, err := svc.GetTask(context.Background(), expected.ID)
	assert.NoError(t, err)
	assert.Equal(t, expected, task)
}
