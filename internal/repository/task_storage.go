package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/veranemoloko/url-downloader/internal/domain"
	errpkg "github.com/veranemoloko/url-downloader/internal/errors"
)

// TaskStorage provides in-memory and file-based storage for tasks.
type TaskStorage struct {
	mu    sync.RWMutex
	tasks map[uuid.UUID]*domain.Task
	file  string
}

// NewTaskStorage creates a new TaskStorage and loads tasks from the file if it exists.
func NewTaskStorage(filePath string) (*TaskStorage, error) {
	repo := &TaskStorage{
		tasks: make(map[uuid.UUID]*domain.Task),
		file:  filepath.Clean(filePath),
	}

	if err := repo.restoreTasks(); err != nil {
		return nil, fmt.Errorf("failed to load state from file: %w", err)
	}

	slog.Info("File repository initialized", "file_path", repo.file, "tasks_count", len(repo.tasks))
	return repo, nil
}

func (r *TaskStorage) restoreTasks() error {
	if isFileNotExist(r.file) {
		slog.Info("State file does not exist, starting with empty state", "file_path", r.file)
		return nil
	}

	data, err := os.ReadFile(r.file)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	if len(data) == 0 {
		slog.Warn("State file is empty")
		return nil
	}

	var tasks []*domain.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return fmt.Errorf("failed to unmarshal state file: %w", err)
	}

	for _, task := range tasks {
		r.tasks[task.ID] = task
	}

	slog.Info("State loaded from file", "tasks_count", len(tasks), "file_path", r.file)
	return nil
}

func isFileNotExist(filePath string) bool {
	_, err := os.Stat(filePath)
	return os.IsNotExist(err)
}

func (r *TaskStorage) persistTasks() error {
	r.mu.RLock()
	tasks := make([]*domain.Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}
	r.mu.RUnlock()

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	tempFile := r.file + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	if err := os.Rename(tempFile, r.file); err != nil {
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	slog.Debug("State saved to file", "tasks_count", len(tasks), "file_path", r.file)
	return nil
}

// CreateTask adds a new task and persists it to the file.
func (r *TaskStorage) CreateTask(ctx context.Context, task *domain.Task) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	r.tasks[task.ID] = task
	r.mu.Unlock()

	if err := r.persistTasks(); err != nil {
		return fmt.Errorf("failed to save state after creating task: %w", err)
	}

	slog.Debug("Task created and saved", "task_id", task.ID)
	return nil
}

// GetTask retrieves a task by ID.
func (r *TaskStorage) GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	task, exists := r.tasks[id]
	r.mu.RUnlock()

	if !exists {
		return nil, errpkg.ErrTaskNotFound
	}
	return task, nil
}

// UpdateTask updates an existing task and persists it to the file.
func (r *TaskStorage) UpdateTask(ctx context.Context, task *domain.Task) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	task.UpdatedAt = time.Now()
	r.tasks[task.ID] = task
	r.mu.Unlock()

	if err := r.persistTasks(); err != nil {
		return fmt.Errorf("failed to save state after updating task: %w", err)
	}

	slog.Debug("Task updated and saved", "task_id", task.ID, "status", task.Status)
	return nil
}

// GetTasksByStatus returns all tasks with the specified status.
func (r *TaskStorage) GetTasksByStatus(ctx context.Context, status domain.TaskStatus) ([]*domain.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	var filtered []*domain.Task
	for _, task := range r.tasks {
		if task.Status == status {
			filtered = append(filtered, task)
		}
	}
	r.mu.RUnlock()

	return filtered, nil
}
