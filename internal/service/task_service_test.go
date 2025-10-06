package service

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/veranemoloko/url-downloader/internal/domain"
	"github.com/veranemoloko/url-downloader/internal/storage"
	"github.com/veranemoloko/url-downloader/internal/worker"
)

func makeTempDir(t *testing.T, prefix string) string {
	t.Helper()
	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

func waitFor(t *testing.T, timeout time.Duration, check func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timeout waiting condition")
}

func TestTaskService_CreateTask_ProcessCompletes(t *testing.T) {
	taskDir := makeTempDir(t, "taskservice_tasks_*")
	downloadDir := makeTempDir(t, "taskservice_downloads_*")

	taskStorage, err := storage.NewTaskStorage(taskDir)
	if err != nil {
		t.Fatalf("NewTaskStorage error: %v", err)
	}
	fileStorage := storage.NewFileStorage(downloadDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/a":
			if _, err := io.WriteString(w, "AAA"); err != nil {
				t.Fatalf("failed to write response for /a: %v", err)
			}
		case "/b":
			if _, err := io.WriteString(w, "BBB"); err != nil {
				t.Fatalf("failed to write response for /b: %v", err)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	logger := newTestLogger()
	wrk := worker.NewDownloadWorker(fileStorage, logger)
	svc := NewTaskService(taskStorage, fileStorage, wrk, logger)

	task, err := svc.CreateTask([]string{server.URL + "/a", server.URL + "/b"})
	if err != nil {
		t.Fatalf("CreateTask error: %v", err)
	}

	waitFor(t, 5*time.Second, func() bool {
		got, err := taskStorage.Get(task.ID)
		if err != nil {
			return false
		}
		return got.Status == domain.StatusCompleted || got.Status == domain.StatusFailed
	})

	final, err := taskStorage.Get(task.ID)
	if err != nil {
		t.Fatalf("failed to get final task: %v", err)
	}

	if final.Status != domain.StatusCompleted {
		t.Fatalf("expected status completed, got %s", final.Status)
	}

	if len(final.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(final.Results))
	}

	for _, r := range final.Results {
		if !r.Success {
			t.Fatalf("expected successful download for %s, got %+v", r.URL, r)
		}
		fp := filepath.Join(downloadDir, r.FileName)
		if _, err := os.Stat(fp); err != nil {
			t.Fatalf("expected downloaded file %s to exist: %v", fp, err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := svc.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown error: %v", err)
	}
}
