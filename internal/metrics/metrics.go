package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TasksCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "url_downloader_tasks_created_total",
		Help: "Total number of tasks created",
	})

	TasksCompleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "url_downloader_tasks_completed_total",
		Help: "Total number of tasks completed",
	})

	TasksFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "url_downloader_tasks_failed_total",
		Help: "Total number of tasks failed",
	})

	DownloadsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "url_downloader_downloads_total",
		Help: "Total number of download attempts",
	})

	DownloadsSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "url_downloader_downloads_success_total",
		Help: "Total number of successful downloads",
	})

	DownloadsFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "url_downloader_downloads_failed_total",
		Help: "Total number of failed downloads",
	})

	DownloadDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "url_downloader_download_duration_seconds",
		Help:    "Download duration in seconds",
		Buckets: prometheus.DefBuckets,
	})

	DownloadBytes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "url_downloader_download_bytes_total",
		Help: "Total bytes downloaded",
	})
)
