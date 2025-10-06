package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/veranemoloko/url-downloader/internal/domain"
)

// TaskStorage provides thread-safe storage and persistence for download tasks.
type TaskStorage struct {
	mu    sync.RWMutex
	dir   string
	tasks map[string]*domain.Task
}

// NewTaskStorage creates a new TaskStorage, loading existing tasks from the specified directory.
func NewTaskStorage(dir string) (*TaskStorage, error) {
	storage := &TaskStorage{
		dir:   dir,
		tasks: make(map[string]*domain.Task),
	}

	if err := storage.loadTasks(); err != nil {
		return nil, fmt.Errorf("load tasks: %w", err)
	}

	return storage, nil
}

func (s *TaskStorage) loadTasks() error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read dir: %w", err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".json" {
			data, err := os.ReadFile(filepath.Join(s.dir, entry.Name()))
			if err != nil {
				return fmt.Errorf("read task file: %w", err)
			}

			var task domain.Task
			if err := json.Unmarshal(data, &task); err != nil {
				return fmt.Errorf("unmarshal task: %w", err)
			}

			s.tasks[task.ID] = &task
		}
	}

	return nil
}

// Save stores or updates a task in memory and persists it to disk.
func (s *TaskStorage) Save(task *domain.Task) error {
	s.mu.Lock()
	s.tasks[task.ID] = task
	s.mu.Unlock()

	return s.persist(task)
}

// Get retrieves a task by its ID. Returns an error if the task does not exist.
func (s *TaskStorage) Get(id string) (*domain.Task, error) {
	s.mu.RLock()
	task, exists := s.tasks[id]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("task not found")
	}

	copyTask := *task
	return &copyTask, nil
}

// GetAll returns a slice of all tasks currently stored in memory.
func (s *TaskStorage) GetAll() []*domain.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*domain.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		copyTask := *task
		tasks = append(tasks, &copyTask)
	}
	return tasks
}

func (s *TaskStorage) persist(task *domain.Task) error {
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal task: %w", err)
	}

	filename := filepath.Join(s.dir, task.ID+".json")
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("write task file: %w", err)
	}

	return nil
}
