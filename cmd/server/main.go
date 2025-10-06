package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/veranemoloko/url-downloader/internal/api"
	"github.com/veranemoloko/url-downloader/internal/config"
	"github.com/veranemoloko/url-downloader/internal/domain"
	"github.com/veranemoloko/url-downloader/internal/service"
	"github.com/veranemoloko/url-downloader/internal/storage"
	"github.com/veranemoloko/url-downloader/internal/worker"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	taskStorage, err := storage.NewTaskStorage(cfg.TaskDir)
	if err != nil {
		logger.Error("failed to initialize task storage", "error", err, "task_dir", cfg.TaskDir)
		os.Exit(1)
	}
	logger.Info("task storage initialized", "task_dir", cfg.TaskDir)

	fileStorage := storage.NewFileStorage(cfg.DownloadDir)
	downloadWorker := worker.NewDownloadWorker(fileStorage, logger)

	taskService := service.NewTaskService(taskStorage, fileStorage, downloadWorker, logger)
	logger.Info("services initialized")

	restoredCount, err := restoreInProgressTasks(taskService, taskStorage, logger)
	if err != nil {
		logger.Error("failed to restore in-progress tasks", "error", err)
	} else if restoredCount > 0 {
		logger.Info("tasks restored after restart", "count", restoredCount)
	}

	router := chi.NewRouter()

	router.Use(slogMiddleware(logger))
	router.Use(middleware.Recoverer)

	taskHandler := api.NewTaskHandler(taskService)
	taskHandler.RegisterRoutes(router)

	server := &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)

	go func() {
		logger.Info("starting HTTP server", "address", cfg.ServerAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", "error", err)
			serverErr <- err
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("application started successfully", "address", cfg.ServerAddress)

	select {
	case err := <-serverErr:
		logger.Error("server runtime error", "error", err)
	case <-done:
		logger.Info("received shutdown signal")
	}

	logger.Info("initiating graceful shutdown")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Info("shutting down HTTP server")
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown failed", "error", err)
	} else {
		logger.Info("HTTP server stopped gracefully")
	}

	logger.Info("shutting down task service")
	if err := taskService.Shutdown(shutdownCtx); err != nil {
		logger.Error("task service shutdown failed", "error", err)
	} else {
		logger.Info("task service stopped gracefully")
	}

	logger.Info("application shutdown completed")
}

func setupLogger(logLevel string) *slog.Logger {
	switch logLevel {
	case "DEBUG":
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	case "WARN":
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelWarn,
		}))
	case "ERROR":
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelError,
		}))
	default:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}
}

func slogMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)
			duration := time.Since(start)
			status := ww.Status()
			if status >= 500 {
				logger.Warn("HTTP request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", status,
					"duration_ms", duration.Milliseconds(),
					"bytes", ww.BytesWritten(),
					"user_agent", r.UserAgent(),
					"remote_addr", r.RemoteAddr,
				)
			} else {
				logger.Info("HTTP request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", status,
					"duration_ms", duration.Milliseconds(),
					"bytes", ww.BytesWritten(),
					"user_agent", r.UserAgent(),
					"remote_addr", r.RemoteAddr,
				)
			}
		})
	}
}

func restoreInProgressTasks(service *service.TaskService, storage *storage.TaskStorage, logger *slog.Logger) (int, error) {
	tasks := storage.GetAll()
	restoredCount := 0

	for _, task := range tasks {
		if task.Status == domain.StatusInProgress {
			logger.Info("restoring in-progress task",
				"task_id", task.ID,
				"urls_count", len(task.URLs),
			)

			task.Status = domain.StatusPending
			if err := storage.Save(task); err != nil {
				return restoredCount, err
			}

			restoredCount++

			go func(t *domain.Task) {
				if err := service.ProcessTask(context.Background(), t); err != nil {
					logger.Error("failed to process restored task",
						"error", err,
						"task_id", t.ID,
					)
				} else {
					logger.Info("restored task processing completed", "task_id", t.ID)
				}
			}(task)
		}
	}

	return restoredCount, nil
}
