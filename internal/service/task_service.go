package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/veranemoloko/url-downloader/internal/domain"
	"github.com/veranemoloko/url-downloader/internal/storage"
	"github.com/veranemoloko/url-downloader/internal/worker"
)

type TaskService struct {
	taskStorage  *storage.TaskStorage
	fileStorage  *storage.FileStorage
	worker       *worker.DownloadWorker
	eventChan    chan domain.TaskEvent
	logger       *slog.Logger
	wg           sync.WaitGroup
	shutdownChan chan struct{}
}

func NewTaskService(
	taskStorage *storage.TaskStorage,
	fileStorage *storage.FileStorage,
	worker *worker.DownloadWorker,
	logger *slog.Logger,
) *TaskService {
	service := &TaskService{
		taskStorage:  taskStorage,
		fileStorage:  fileStorage,
		worker:       worker,
		eventChan:    make(chan domain.TaskEvent, 100),
		logger:       logger,
		shutdownChan: make(chan struct{}),
	}

	service.wg.Add(1)
	go service.eventProcessor()

	return service
}

func (s *TaskService) CreateTask(urls []string) (*domain.Task, error) {
	task := &domain.Task{
		ID:        generateID(),
		URLs:      urls,
		Status:    domain.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	select {
	case s.eventChan <- domain.TaskEvent{
		Type:   domain.EventCreateTask,
		TaskID: task.ID,
		Task:   task,
	}:
		s.logger.Info("task created",
			"task_id", task.ID,
			"urls_count", len(urls),
		)
		return task, nil
	case <-s.shutdownChan:
		return nil, fmt.Errorf("service is shutting down")
	}
}

func (s *TaskService) GetTask(id string) (*domain.Task, error) {
	return s.taskStorage.Get(id)
}

func (s *TaskService) ProcessTask(ctx context.Context, task *domain.Task) error {
	s.logger.Info("start processing task",
		"task_id", task.ID,
		"urls_count", len(task.URLs),
	)

	inProgressStatus := domain.StatusInProgress
	update := &domain.TaskUpdate{
		Status: &inProgressStatus,
	}

	select {
	case s.eventChan <- domain.TaskEvent{
		Type:    domain.EventUpdateTask,
		TaskID:  task.ID,
		Updates: update,
	}:
	case <-s.shutdownChan:
		return fmt.Errorf("service is shutting down")
	case <-ctx.Done():
		return ctx.Err()
	}

	downloadCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		select {
		case <-s.shutdownChan:
			cancel()
		case <-downloadCtx.Done():
		}
	}()

	results, err := s.worker.DownloadTask(downloadCtx, task)

	select {
	case <-s.shutdownChan:
		return fmt.Errorf("service shutdown during download")
	default:

		update = &domain.TaskUpdate{
			Results: results,
		}

		if err != nil {
			status := domain.StatusFailed
			update.Status = &status

			s.logger.Error("task processing failed",
				"task_id", task.ID,
				"error", err,
			)
		} else {
			allSuccess := true
			successCount := 0
			for _, result := range results {
				if result.Success {
					successCount++
				} else {
					allSuccess = false
				}
			}

			if allSuccess {
				status := domain.StatusCompleted
				update.Status = &status

				s.logger.Info("task completed successfully",
					"task_id", task.ID,
					"successful_downloads", successCount,
					"total_downloads", len(results),
				)
			} else {
				status := domain.StatusFailed
				update.Status = &status

				s.logger.Warn("task completed with failures",
					"task_id", task.ID,
					"successful_downloads", successCount,
					"failed_downloads", len(results)-successCount,
					"total_downloads", len(results),
				)
			}
		}

		select {
		case s.eventChan <- domain.TaskEvent{
			Type:    domain.EventUpdateTask,
			TaskID:  task.ID,
			Updates: update,
		}:
		case <-s.shutdownChan:
			return fmt.Errorf("service shutdown during state update")
		}
	}

	return err
}

func (s *TaskService) eventProcessor() {
	defer s.wg.Done()

	for {
		select {
		case event, ok := <-s.eventChan:
			if !ok {
				return
			}

			switch event.Type {
			case domain.EventCreateTask:
				if err := s.taskStorage.Save(event.Task); err != nil {
					s.logger.Error("failed to save task",
						"error", err,
						"task_id", event.TaskID,
					)
				} else {
					s.logger.Debug("task saved to storage",
						"task_id", event.TaskID,
					)
				}

				s.wg.Add(1)
				go func(task *domain.Task) {
					defer s.wg.Done()
					if err := s.ProcessTask(context.Background(), task); err != nil {
						s.logger.Error("failed to process task",
							"error", err,
							"task_id", task.ID,
						)
					}
				}(event.Task)

			case domain.EventUpdateTask:
				task, err := s.taskStorage.Get(event.TaskID)
				if err != nil {
					s.logger.Error("failed to get task for update",
						"error", err,
						"task_id", event.TaskID,
					)
					continue
				}

				if event.Updates.Status != nil {
					task.Status = *event.Updates.Status
				}
				if event.Updates.Results != nil {
					task.Results = event.Updates.Results
				}
				task.UpdatedAt = time.Now()

				if err := s.taskStorage.Save(task); err != nil {
					s.logger.Error("failed to save task update",
						"error", err,
						"task_id", event.TaskID,
						"status", task.Status,
					)
				} else {
					s.logger.Debug("task state updated",
						"task_id", event.TaskID,
						"status", task.Status,
					)
				}
			}

		case <-s.shutdownChan:
			for {
				select {
				case event := <-s.eventChan:
					if event.Type == domain.EventUpdateTask {
						if task, err := s.taskStorage.Get(event.TaskID); err == nil {
							if event.Updates.Status != nil {
								task.Status = *event.Updates.Status
							}
							task.UpdatedAt = time.Now()
							s.taskStorage.Save(task)
						}
					}
				default:
					return
				}
			}
		}
	}
}

func (s *TaskService) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down task service")

	close(s.shutdownChan)
	close(s.eventChan)

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("task service shutdown completed")
		return nil
	case <-ctx.Done():
		s.logger.Warn("task service shutdown timed out")
		return ctx.Err()
	}
}

func generateID() string {
	return uuid.New().String()
}
