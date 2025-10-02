package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/veranemoloko/url-downloader/internal/domain"
)

type mockTaskService struct{}

func (m *mockTaskService) CreateTask(ctx context.Context, req *domain.CreateTaskRequest) (*domain.Task, error) {
	return &domain.Task{ID: uuid.New(), Status: domain.TaskStatusPending, URLs: req.URLs}, nil
}
func (m *mockTaskService) GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	return &domain.Task{ID: id, Status: domain.TaskStatusCompleted, URLs: []string{"http://a"}}, nil
}

func TestTaskHandler_CreateTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	handler := NewTaskHandler(&mockTaskService{}, logger)

	body, _ := json.Marshal(domain.CreateTaskRequest{URLs: []string{"http://a"}})
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateTask(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var data map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&data)
	assert.Contains(t, data, "task_id")
}

func TestTaskHandler_GetTask(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	handler := NewTaskHandler(&mockTaskService{}, logger)

	id := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/tasks/"+id.String(), nil)

	r := chi.NewRouter()
	r.Get("/tasks/{taskID}", handler.GetTask)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var data domain.TaskResponse
	_ = json.NewDecoder(resp.Body).Decode(&data)
	assert.Equal(t, id, data.ID)
}
