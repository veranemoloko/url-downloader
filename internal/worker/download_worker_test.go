package worker

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/veranemoloko/url-downloader/internal/domain"
	"github.com/veranemoloko/url-downloader/internal/storage"
)

func makeTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "downloadworker_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestDownloadWorker_DownloadURL_FullDownload(t *testing.T) {
	dir := makeTempDir(t)
	fs := storage.NewFileStorage(dir)
	logger := newTestLogger()
	worker := NewDownloadWorker(fs, logger)

	wantContent := "hello world"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "11")
		w.WriteHeader(http.StatusOK)
		if _, err := io.WriteString(w, wantContent); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	taskID := "task1"

	result, err := worker.DownloadURL(ctx, server.URL, taskID)
	if err != nil {
		t.Fatalf("DownloadURL error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected Success=true, got false")
	}
	if result.BytesRead != int64(len(wantContent)) {
		t.Errorf("expected BytesRead=%d, got %d", len(wantContent), result.BytesRead)
	}

	filePath := filepath.Join(dir, result.FileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(data) != wantContent {
		t.Errorf("expected file content %q, got %q", wantContent, string(data))
	}
}

func TestDownloadWorker_DownloadURL_Resume(t *testing.T) {
	dir := makeTempDir(t)
	fs := storage.NewFileStorage(dir)
	logger := newTestLogger()
	worker := NewDownloadWorker(fs, logger)

	taskID := "task2"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeHeader := r.Header.Get("Range")
		if strings.HasPrefix(rangeHeader, "bytes=3-") {
			w.WriteHeader(http.StatusPartialContent)
			if _, err := io.WriteString(w, "lo world"); err != nil {
				t.Fatalf("failed to write partial response: %v", err)
			}
		} else {
			w.WriteHeader(http.StatusOK)
			if _, err := io.WriteString(w, "hello world"); err != nil {
				t.Fatalf("failed to write full response: %v", err)
			}
		}
	}))
	defer server.Close()

	fileName := worker.generateFilename(server.URL, taskID)
	filePath := filepath.Join(dir, fileName)

	if err := os.WriteFile(filePath, []byte("hel"), 0644); err != nil {
		t.Fatalf("failed to create partial file: %v", err)
	}

	ctx := context.Background()
	result, err := worker.DownloadURL(ctx, server.URL, taskID)
	if err != nil {
		t.Fatalf("DownloadURL resume error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected Success=true, got false")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("expected resumed content 'hello world', got %q", string(data))
	}
}

func TestDownloadWorker_DownloadURL_HTTPError(t *testing.T) {
	dir := makeTempDir(t)
	fs := storage.NewFileStorage(dir)
	logger := newTestLogger()
	worker := NewDownloadWorker(fs, logger)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	ctx := context.Background()
	taskID := "task3"

	result, err := worker.DownloadURL(ctx, server.URL, taskID)
	if err == nil {
		t.Errorf("expected error for 500 response, got nil")
	}
	if result.Success {
		t.Errorf("expected Success=false for 500 response")
	}
}

func TestDownloadWorker_DownloadTask_Multiple(t *testing.T) {
	dir := makeTempDir(t)
	fs := storage.NewFileStorage(dir)
	logger := newTestLogger()
	worker := NewDownloadWorker(fs, logger)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/a":
			if _, err := io.WriteString(w, "aaa"); err != nil {
				t.Fatalf("failed to write response for /a: %v", err)
			}
		case "/b":
			if _, err := io.WriteString(w, "bbb"); err != nil {
				t.Fatalf("failed to write response for /b: %v", err)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	task := &domain.Task{
		ID:   "multi",
		URLs: []string{server.URL + "/a", server.URL + "/b"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	results, err := worker.DownloadTask(ctx, task)
	if err != nil {
		t.Fatalf("DownloadTask error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for _, r := range results {
		if !r.Success {
			t.Errorf("expected all downloads to succeed, got failed: %+v", r)
		}
	}
}
