package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type FileStorage struct {
	dir string
}

func NewFileStorage(dir string) *FileStorage {
	return &FileStorage{dir: dir}
}

func (s *FileStorage) CreateFile(filename string) (*os.File, error) {
	filepath := filepath.Join(s.dir, filename)
	return os.Create(filepath)
}

func (s *FileStorage) OpenFile(filename string, flags int) (*os.File, error) {
	filepath := filepath.Join(s.dir, filename)
	return os.OpenFile(filepath, flags, 0644)
}

func (s *FileStorage) FileExists(filename string) bool {
	filepath := filepath.Join(s.dir, filename)
	_, err := os.Stat(filepath)
	return err == nil
}

func (s *FileStorage) GetFileSize(filename string) (int64, error) {
	filepath := filepath.Join(s.dir, filename)
	info, err := os.Stat(filepath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (s *FileStorage) WriteFile(filename string, data []byte) error {
	filepath := filepath.Join(s.dir, filename)
	return os.WriteFile(filepath, data, 0644)
}

func (s *FileStorage) CopyFile(src io.Reader, dstFilename string) (int64, error) {
	dst, err := s.CreateFile(dstFilename)
	if err != nil {
		return 0, fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	return io.Copy(dst, src)
}
