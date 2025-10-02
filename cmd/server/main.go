package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	h "github.com/veranemoloko/url-downloader/internal/api/http"
	cfgpkg "github.com/veranemoloko/url-downloader/internal/config"
	repo "github.com/veranemoloko/url-downloader/internal/repository"
	svc "github.com/veranemoloko/url-downloader/internal/service"
)

func main() {

	cfg, err := cfgpkg.Load()
	if err != nil {
		var pathErr *os.PathError
		if errors.As(err, &pathErr) {
			slog.Error("configuration file not found", "error", err)
		} else {
			slog.Error("failed to load configuration", "error", err)
		}
		os.Exit(1)
	}

	cfgpkg.SetupLogger(cfg)
	slog.Info("configuration loaded successfully")

	taskStorage, err := repo.NewTaskStorage(cfg.StateFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Error("state file does not exist", "error", err)
		} else {
			slog.Error("failed to initialize file repository", "error", err)
		}
		os.Exit(1)
	}

	downloadService := svc.NewDownloadService(taskStorage, cfg)
	taskService := svc.NewTaskService(taskStorage, downloadService)

	if err := downloadService.RecoverPendingTasks(context.Background()); err != nil {
		slog.Error("failed to recover pending tasks", "error", err)
	}

	router := h.NewRouter(taskService, slog.Default())
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      router,
		ReadTimeout:  cfg.HTTPTimeout,
		WriteTimeout: cfg.HTTPTimeout,
		IdleTimeout:  cfg.HTTPTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("server starting", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown failed", "error", err)
	} else {
		slog.Info("server stopped gracefully")
	}
}
