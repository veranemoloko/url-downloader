package http

import (
	"context"
	"encoding/json"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/veranemoloko/url-downloader/internal/domain"
)

// TaskServiceI defines the interface for task-related business logic.
type TaskServiceI interface {
	CreateTask(ctx context.Context, req *domain.CreateTaskRequest) (*domain.Task, error)
	GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error)
}

// TaskHandler handles HTTP requests for tasks.
type TaskHandler struct {
	taskService TaskServiceI
	validator   *validator.Validate
	logger      *slog.Logger
}

// NewTaskHandler creates a new TaskHandler with the provided service and logger.
func NewTaskHandler(taskService TaskServiceI, logger *slog.Logger) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
		validator:   validator.New(),
		logger:      logger,
	}
}

// CreateTask handles the HTTP POST /tasks request to create a new task.
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req domain.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", "error", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.logger.Warn("validation failed", "error", err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := h.taskService.CreateTask(ctx, &req)
	if err != nil {
		h.logger.Error("failed to create task", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info("task created", "task_id", task.ID, "urls_count", len(task.URLs))

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"task_id": task.ID,
	})
}

// GetTask handles the HTTP GET /tasks/{taskID} request to fetch a task by ID.
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	taskIDStr := chi.URLParam(r, "taskID")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task ID")
		return
	}

	task, err := h.taskService.GetTask(ctx, taskID)
	if err != nil {
		h.logger.Error("failed to get task", "task_id", taskID, "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if task == nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	response := domain.TaskResponse{
		ID:        task.ID,
		Status:    task.Status,
		Downloads: task.Downloads,
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
	}

	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}
