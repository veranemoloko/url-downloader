package worker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/veranemoloko/url-downloader/internal/domain"
	"github.com/veranemoloko/url-downloader/internal/storage"
	"golang.org/x/sync/errgroup"
)

type DownloadWorker struct {
	fileStorage *storage.FileStorage
	httpClient  *http.Client
	logger      *slog.Logger
}

func NewDownloadWorker(fileStorage *storage.FileStorage, logger *slog.Logger) *DownloadWorker {
	return &DownloadWorker{
		fileStorage: fileStorage,
		httpClient: &http.Client{
			Timeout: 30 * time.Minute,
		},
		logger: logger,
	}
}

func (w *DownloadWorker) DownloadURL(ctx context.Context, url string, taskID string) (domain.DownloadResult, error) {
	result := domain.DownloadResult{
		URL:     url,
		Success: false,
	}

	filename := w.generateFilename(url, taskID)
	result.FileName = filename

	var existingSize int64 = 0
	if w.fileStorage.FileExists(filename) {
		size, err := w.fileStorage.GetFileSize(filename)
		if err == nil {
			existingSize = size
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		result.Error = fmt.Sprintf("create request: %v", err)
		w.logger.Error("download failed",
			"url", url,
			"error", err,
		)
		return result, err
	}

	if existingSize > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", existingSize))
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("do request: %v", err)
		w.logger.Error("download failed",
			"url", url,
			"error", err,
		)
		return result, err
	}
	defer resp.Body.Close()

	if existingSize > 0 && resp.StatusCode != http.StatusPartialContent {
		existingSize = 0
	}

	var file *os.File
	var flags int

	if existingSize > 0 {
		flags = os.O_WRONLY | os.O_APPEND
		file, err = w.fileStorage.OpenFile(filename, flags)
		if err != nil {
			result.Error = fmt.Sprintf("open file for append: %v", err)
			w.logger.Error("download failed",
				"url", url,
				"error", err,
			)
			return result, err
		}
	} else {
		file, err = w.fileStorage.CreateFile(filename)
		if err != nil {
			result.Error = fmt.Sprintf("create file: %v", err)
			w.logger.Error("download failed",
				"url", url,
				"error", err,
			)
			return result, err
		}
	}
	defer file.Close()

	bytesRead, err := w.copyWithContext(ctx, file, resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("copy data: %v", err)
		w.logger.Error("download failed",
			"url", url,
			"error", err,
		)
		return result, err
	}

	totalBytes := existingSize + bytesRead
	result.BytesRead = totalBytes
	result.Success = true

	return result, nil
}

func (w *DownloadWorker) copyWithContext(ctx context.Context, dst *os.File, src io.Reader) (int64, error) {
	buf := make([]byte, 32*1024)
	var total int64

	for {
		select {
		case <-ctx.Done():
			return total, ctx.Err()
		default:
			nr, err := src.Read(buf)
			if nr > 0 {
				nw, err := dst.Write(buf[0:nr])
				if nw > 0 {
					total += int64(nw)
				}
				if err != nil {
					return total, err
				}
				if nr != nw {
					return total, io.ErrShortWrite
				}
			}
			if err != nil {
				if err == io.EOF {
					return total, nil
				}
				return total, err
			}
		}
	}
}

func (w *DownloadWorker) DownloadTask(ctx context.Context, task *domain.Task) ([]domain.DownloadResult, error) {
	results := make([]domain.DownloadResult, len(task.URLs))
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(5)

	for i, url := range task.URLs {
		i, url := i, url
		g.Go(func() error {
			result, err := w.DownloadURL(ctx, url, task.ID)
			results[i] = result
			return err
		})
	}

	if err := g.Wait(); err != nil {
		w.logger.Error("task download failed",
			"task_id", task.ID,
			"error", err,
		)
		return results, fmt.Errorf("download task: %w", err)
	}

	return results, nil
}

func (w *DownloadWorker) generateFilename(url, taskID string) string {
	return fmt.Sprintf("%s_%x", taskID, url)
}
