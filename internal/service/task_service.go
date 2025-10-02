package service

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/google/uuid"

	"github.com/veranemoloko/url-downloader/internal/domain"
	"github.com/veranemoloko/url-downloader/internal/metrics"
	repo "github.com/veranemoloko/url-downloader/internal/repository"
)

type TaskService struct {
	taskRepo        repo.TaskRepo
	downloadService *DownloadService
}

func NewTaskService(taskRepo repo.TaskRepo, downloadService *DownloadService) *TaskService {
	return &TaskService{
		taskRepo:        taskRepo,
		downloadService: downloadService,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, req *domain.CreateTaskRequest) (*domain.Task, error) {

	task := &domain.Task{
		ID:        uuid.New(),
		Status:    domain.TaskStatusPending,
		URLs:      req.URLs,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, url := range req.URLs {
		task.Downloads = append(task.Downloads, domain.DownloadItem{
			URL:    url,
			Status: domain.DownloadStatusPending,
		})
	}

	if err := s.taskRepo.CreateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	metrics.TasksCreated.Inc()
	slog.Info("task created successfully", "task_id", task.ID, "urls_count", len(task.URLs))

	if s.downloadService != nil {
		go func() {
			if err := s.downloadService.ProcessTask(context.Background(), task); err != nil {
				slog.Error("failed to process task", "task_id", task.ID, "error", err)
				metrics.TasksFailed.Inc()
				task.Status = domain.TaskStatusFailed
				task.UpdatedAt = time.Now()
				if updateErr := s.taskRepo.UpdateTask(context.Background(), task); updateErr != nil {
					slog.Error("failed to update task status to failed", "task_id", task.ID, "error", updateErr)
				}
			} else {
				metrics.TasksCompleted.Inc()
			}
		}()
	}

	return task, nil
}

func (s *TaskService) GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	task, err := s.taskRepo.GetTask(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}
