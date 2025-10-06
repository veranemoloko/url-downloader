package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func makeTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "filestorage_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

func TestFileStorage_CreateAndExists(t *testing.T) {
	dir := makeTempDir(t)
	fs := NewFileStorage(dir)

	f, err := fs.CreateFile("test.txt")
	if err != nil {
		t.Fatalf("CreateFile error: %v", err)
	}
	f.Close()

	if !fs.FileExists("test.txt") {
		t.Errorf("expected file to exist after creation")
	}
}

func TestFileStorage_WriteAndSize(t *testing.T) {
	dir := makeTempDir(t)
	fs := NewFileStorage(dir)

	data := []byte("hello world")
	if err := fs.WriteFile("data.txt", data); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	size, err := fs.GetFileSize("data.txt")
	if err != nil {
		t.Fatalf("GetFileSize error: %v", err)
	}

	if size != int64(len(data)) {
		t.Errorf("expected size %d, got %d", len(data), size)
	}
}

func TestFileStorage_OpenFileAppend(t *testing.T) {
	dir := makeTempDir(t)
	fs := NewFileStorage(dir)

	if err := fs.WriteFile("append.txt", []byte("part1")); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	f, err := fs.OpenFile("append.txt", os.O_WRONLY|os.O_APPEND)
	if err != nil {
		t.Fatalf("OpenFile error: %v", err)
	}

	if _, err := f.Write([]byte("part2")); err != nil {
		t.Fatalf("append write error: %v", err)
	}
	f.Close()

	content, err := os.ReadFile(filepath.Join(dir, "append.txt"))
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	if string(content) != "part1part2" {
		t.Errorf("expected 'part1part2', got %q", string(content))
	}
}

func TestFileStorage_CopyFile(t *testing.T) {
	dir := makeTempDir(t)
	fs := NewFileStorage(dir)

	srcData := []byte("copy test content")
	srcReader := bytes.NewReader(srcData)

	n, err := fs.CopyFile(srcReader, "copied.txt")
	if err != nil {
		t.Fatalf("CopyFile error: %v", err)
	}

	if n != int64(len(srcData)) {
		t.Errorf("expected copied bytes %d, got %d", len(srcData), n)
	}

	readBack, err := os.ReadFile(filepath.Join(dir, "copied.txt"))
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}

	if !bytes.Equal(readBack, srcData) {
		t.Errorf("copied content mismatch: got %q, want %q", readBack, srcData)
	}
}

func TestFileStorage_FileExistsFalse(t *testing.T) {
	dir := makeTempDir(t)
	fs := NewFileStorage(dir)

	if fs.FileExists("no_such_file.txt") {
		t.Errorf("expected FileExists to return false for non-existing file")
	}
}
