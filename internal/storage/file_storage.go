package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileStorage provides methods to manage files in a specific directory.
type FileStorage struct {
	dir string
}

// NewFileStorage creates a new FileStorage instance with the given directory.
func NewFileStorage(dir string) *FileStorage {
	return &FileStorage{dir: dir}
}

// CreateFile creates a new file with the given filename in the storage directory.
func (s *FileStorage) CreateFile(filename string) (*os.File, error) {
	filepath := filepath.Join(s.dir, filename)
	return os.Create(filepath)
}

// OpenFile opens an existing file with the specified flags (e.g., read, write).
func (s *FileStorage) OpenFile(filename string, flags int) (*os.File, error) {
	filepath := filepath.Join(s.dir, filename)
	return os.OpenFile(filepath, flags, 0644)
}

// FileExists checks whether a file exists in the storage directory.
func (s *FileStorage) FileExists(filename string) bool {
	filepath := filepath.Join(s.dir, filename)
	_, err := os.Stat(filepath)
	return err == nil
}

// GetFileSize returns the size of the file in bytes.
func (s *FileStorage) GetFileSize(filename string) (int64, error) {
	filepath := filepath.Join(s.dir, filename)
	info, err := os.Stat(filepath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// WriteFile writes the given data to a file with the specified filename.
func (s *FileStorage) WriteFile(filename string, data []byte) error {
	filepath := filepath.Join(s.dir, filename)
	return os.WriteFile(filepath, data, 0644)
}

// CopyFile copies data from the provided reader to a file with the specified filename.
// Returns the number of bytes written and any error encountered.
func (s *FileStorage) CopyFile(src io.Reader, dstFilename string) (int64, error) {
	dst, err := s.CreateFile(dstFilename)
	if err != nil {
		return 0, fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	return io.Copy(dst, src)
}
