package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/veranemoloko/url-downloader/internal/domain"
	"github.com/veranemoloko/url-downloader/internal/service"
	"github.com/veranemoloko/url-downloader/internal/validation"
)

var validate = validator.New()

type TaskHandler struct {
	service *service.TaskService
}

func NewTaskHandler(service *service.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

type CreateTaskRequest struct {
	URLs []string `json:"urls" validate:"required,min=1,max=100,dive,required,url"`
}

type TaskResponse struct {
	ID        string                  `json:"id"`
	URLs      []string                `json:"urls"`
	Status    domain.TaskStatus       `json:"status"`
	Results   []domain.DownloadResult `json:"results,omitempty"`
	CreatedAt string                  `json:"created_at"`
	UpdatedAt string                  `json:"updated_at"`
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validation.ValidateURLs(req.URLs); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	task, err := h.service.CreateTask(req.URLs)
	if err != nil {
		sendError(w, "create task failed", http.StatusInternalServerError)
		return
	}

	response := TaskResponse{
		ID:        task.ID,
		URLs:      task.URLs,
		Status:    task.Status,
		Results:   task.Results,
		CreatedAt: task.CreatedAt.Format(time.RFC3339),
		UpdatedAt: task.UpdatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		sendError(w, "task id is required", http.StatusBadRequest)
		return
	}

	task, err := h.service.GetTask(taskID)
	if err != nil {
		sendError(w, "task not found", http.StatusNotFound)
		return
	}

	response := TaskResponse{
		ID:        task.ID,
		URLs:      task.URLs,
		Status:    task.Status,
		Results:   task.Results,
		CreatedAt: task.CreatedAt.Format(time.RFC3339),
		UpdatedAt: task.UpdatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *TaskHandler) RegisterRoutes(router chi.Router) {
	router.Route("/tasks", func(r chi.Router) {
		r.Post("/", h.CreateTask)
		r.Get("/{id}", h.GetTask)
	})
}
