package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"github.com/veranemoloko/url-downloader/internal/domain"
)

type mockTaskService struct{}

func (m *mockTaskService) CreateTask(urls []string) (*domain.Task, error) {
	return &domain.Task{
		ID:        "test-id",
		URLs:      urls,
		Status:    domain.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockTaskService) GetTask(id string) (*domain.Task, error) {
	return &domain.Task{
		ID:        id,
		URLs:      []string{"http://example.com"},
		Status:    domain.StatusCompleted,
		Results:   []domain.DownloadResult{{URL: "http://example.com", Success: true, FileName: "file"}},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func TestTaskHandler_CreateTask(t *testing.T) {
	svc := &mockTaskService{}
	handler := NewTaskHandler(svc)

	reqBody := map[string]interface{}{
		"urls": []string{"http://example.com"},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.CreateTask(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	require.Equal(t, "test-id", resp["id"])
}

func TestTaskHandler_GetTask(t *testing.T) {
	svc := &mockTaskService{}
	handler := NewTaskHandler(svc)

	// создаём route context и подставляем параметр
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-id")

	req := httptest.NewRequest(http.MethodGet, "/tasks/test-id", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.GetTask(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	require.Equal(t, "test-id", resp["id"])
}
