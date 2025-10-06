package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/veranemoloko/url-downloader/internal/domain"
)

func TestTaskStorage_SaveAndGet(t *testing.T) {
	dir := makeTempDir(t)
	storage, err := NewTaskStorage(dir)
	if err != nil {
		t.Fatalf("NewTaskStorage error: %v", err)
	}

	task := &domain.Task{
		ID:     "task1",
		Status: "pending",
		URLs:   []string{"https://example.com"},
	}

	if err := storage.Save(task); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	path := filepath.Join(dir, "task1.json")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file %s to exist, got error: %v", path, err)
	}

	got, err := storage.Get("task1")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}

	if got.ID != task.ID {
		t.Errorf("expected ID %q, got %q", task.ID, got.ID)
	}
	if got.Status != task.Status {
		t.Errorf("expected Status %q, got %q", task.Status, got.Status)
	}
}

func TestTaskStorage_LoadTasks(t *testing.T) {
	dir := makeTempDir(t)

	task := domain.Task{ID: "preloaded", Status: "done", URLs: []string{"http://golang.org"}}
	data, _ := json.MarshalIndent(task, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "preloaded.json"), data, 0644); err != nil {
		t.Fatalf("failed to write preload file: %v", err)
	}

	storage, err := NewTaskStorage(dir)
	if err != nil {
		t.Fatalf("NewTaskStorage error: %v", err)
	}

	got, err := storage.Get("preloaded")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if got.Status != "done" {
		t.Errorf("expected Status 'done', got %q", got.Status)
	}
}

func TestTaskStorage_GetAll(t *testing.T) {
	dir := makeTempDir(t)
	storage, err := NewTaskStorage(dir)
	if err != nil {
		t.Fatalf("NewTaskStorage error: %v", err)
	}

	task1 := &domain.Task{ID: "t1", Status: "pending"}
	task2 := &domain.Task{ID: "t2", Status: "done"}

	if err := storage.Save(task1); err != nil {
		t.Fatalf("Save error: %v", err)
	}
	if err := storage.Save(task2); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	all := storage.GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(all))
	}

	found := map[string]bool{"t1": false, "t2": false}
	for _, tsk := range all {
		found[tsk.ID] = true
	}
	for id, ok := range found {
		if !ok {
			t.Errorf("expected task with ID %s in GetAll result", id)
		}
	}
}

func TestTaskStorage_GetNotFound(t *testing.T) {
	dir := makeTempDir(t)
	storage, err := NewTaskStorage(dir)
	if err != nil {
		t.Fatalf("NewTaskStorage error: %v", err)
	}

	if _, err := storage.Get("missing"); err == nil {
		t.Errorf("expected error for missing task, got nil")
	}
}
