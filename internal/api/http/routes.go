package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/veranemoloko/url-downloader/internal/service"
)

// NewRouter creates a new HTTP router with configured routes, middleware, and handlers.
// It sets up task routes, health check, and Prometheus metrics endpoint.
func NewRouter(taskService *service.TaskService, logger *slog.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	taskHandler := NewTaskHandler(taskService, logger)

	r.Route("/tasks", func(r chi.Router) {
		r.Post("/", taskHandler.CreateTask)
		r.Get("/{taskID}", taskHandler.GetTask)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Handle("/metrics", promhttp.Handler())

	return r
}
