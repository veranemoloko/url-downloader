package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"github.com/veranemoloko/url-downloader/internal/config"
	"github.com/veranemoloko/url-downloader/internal/domain"
	"github.com/veranemoloko/url-downloader/internal/metrics"
	repo "github.com/veranemoloko/url-downloader/internal/repository"
)

type DownloadService struct {
	taskRepo      repo.TaskRepo
	cfg           *config.Config
	downloadQueue chan *downloadJob
	wg            sync.WaitGroup
}

type downloadJob struct {
	taskID    uuid.UUID
	url       string
	itemIndex int
}

// NewDownloadService creates a new DownloadService with a worker pool to process downloads.
func NewDownloadService(taskRepo repo.TaskRepo, cfg *config.Config) *DownloadService {
	service := &DownloadService{
		taskRepo:      taskRepo,
		cfg:           cfg,
		downloadQueue: make(chan *downloadJob, 100),
	}

	client := &http.Client{Timeout: cfg.DownloadTimeout}

	for i := 0; i < cfg.WorkerPoolSize; i++ {
		service.wg.Add(1)
		go func(workerID int) {
			defer service.wg.Done()
			for job := range service.downloadQueue {
				if err := service.processDownload(client, job); err != nil {
					slog.Error("worker failed to process job", "worker_id", workerID, "task_id", job.taskID, "url", job.url, "error", err)
				}
			}
		}(i + 1)
	}

	slog.Info("download service started", "workers", cfg.WorkerPoolSize)
	return service
}

// ProcessTask marks the task as in-progress and enqueues download jobs for all URLs in the task.
func (s *DownloadService) ProcessTask(ctx context.Context, task *domain.Task) error {
	task.Status = domain.TaskStatusInProgress
	task.UpdatedAt = time.Now()

	if err := s.taskRepo.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	slog.Info("processing task", "task_id", task.ID, "urls_count", len(task.URLs))

	for i, url := range task.URLs {
		s.downloadQueue <- &downloadJob{taskID: task.ID, url: url, itemIndex: i}
	}

	return nil
}

func (s *DownloadService) processDownload(client *http.Client, job *downloadJob) error {
	task, err := s.taskRepo.GetTask(context.Background(), job.taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return fmt.Errorf("task not found: %s", job.taskID)
	}

	item := &task.Downloads[job.itemIndex]
	item.Status = domain.DownloadStatusInProgress

	startTime := time.Now()
	filePath, err := s.downloadFile(client, job.url, job.taskID, job.itemIndex)
	duration := time.Since(startTime)

	metrics.DownloadsTotal.Inc()

	if err != nil {
		item.Status = domain.DownloadStatusFailed
		item.Error = err.Error()
		slog.Error("download failed", "task_id", job.taskID, "url", job.url, "error", err)

		metrics.DownloadsFailed.Inc()
	} else {
		item.Status = domain.DownloadStatusCompleted
		item.FilePath = filePath
		slog.Info("download completed", "task_id", job.taskID, "url", job.url, "file_path", filePath)

		metrics.DownloadsSuccess.Inc()

		metrics.DownloadDuration.Observe(duration.Seconds())

		if fileInfo, err := os.Stat(filePath); err == nil {
			metrics.DownloadBytes.Add(float64(fileInfo.Size()))
		}
	}

	if s.isTaskCompleted(task) {
		task.Status = domain.TaskStatusCompleted
		task.UpdatedAt = time.Now()
	}

	return s.taskRepo.UpdateTask(context.Background(), task)
}

func (s *DownloadService) downloadFile(client *http.Client, url string, taskID uuid.UUID, index int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.DownloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	fileName := fmt.Sprintf("%s_%d", taskID.String(), index)
	fileExt := filepath.Ext(url)
	if fileExt == "" {
		fileExt = ".bin"
	}
	filePath := filepath.Join(s.cfg.DownloadDir, fileName+fileExt)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	limitedReader := &io.LimitedReader{R: resp.Body, N: s.cfg.MaxFileSize}
	bytesCopied, err := io.Copy(file, limitedReader)
	if err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to save file: %w", err)
	}
	if limitedReader.N <= 0 {
		os.Remove(filePath)
		return "", fmt.Errorf("file size exceeds limit: %d bytes", s.cfg.MaxFileSize)
	}

	slog.Debug("file downloaded successfully", "task_id", taskID, "url", url, "bytes", bytesCopied, "file_path", filePath)
	return filePath, nil
}

func (s *DownloadService) isTaskCompleted(task *domain.Task) bool {
	for _, download := range task.Downloads {
		if download.Status != domain.DownloadStatusCompleted && download.Status != domain.DownloadStatusFailed {
			return false
		}
	}
	return true
}

// RecoverPendingTasks recovers tasks with Pending or InProgress status and resumes their downloads.
func (s *DownloadService) RecoverPendingTasks(ctx context.Context) error {
	pending, err := s.taskRepo.GetTasksByStatus(ctx, domain.TaskStatusPending)
	if err != nil {
		return fmt.Errorf("failed to get pending tasks: %w", err)
	}

	inProgress, err := s.taskRepo.GetTasksByStatus(ctx, domain.TaskStatusInProgress)
	if err != nil {
		return fmt.Errorf("failed to get in-progress tasks: %w", err)
	}

	tasks := append(pending, inProgress...)

	for _, task := range tasks {
		if err := ctx.Err(); err != nil {
			return err
		}

		if task.Status == domain.TaskStatusInProgress {
			for i := range task.Downloads {
				if task.Downloads[i].Status == domain.DownloadStatusInProgress {
					task.Downloads[i].Status = domain.DownloadStatusPending
					task.Downloads[i].Error = ""
				}
			}

			if err := s.taskRepo.UpdateTask(ctx, task); err != nil {
				slog.Error("failed to recover task", "task_id", task.ID, "error", err)
				continue
			}
		}

		go func(t *domain.Task) {
			if err := s.ProcessTask(context.Background(), t); err != nil {
				slog.Error("failed to process recovered task", "task_id", t.ID, "error", err)
			}
		}(task)
	}

	return nil
}
