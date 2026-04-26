package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	abs, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("resolve storage path: %w", err)
	}

	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, fmt.Errorf("create storage directory: %w", err)
	}

	return &LocalStorage{basePath: abs}, nil
}

// safePath resolves objectName under basePath and rejects traversal attempts.
func (s *LocalStorage) safePath(objectName string) (string, error) {
	if filepath.IsAbs(objectName) {
		return "", fmt.Errorf("path %q escapes storage root", objectName)
	}

	cleaned := filepath.FromSlash(objectName)
	full := filepath.Join(s.basePath, cleaned)
	abs, err := filepath.Abs(full)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}

	if !strings.HasPrefix(abs, s.basePath+string(filepath.Separator)) && abs != s.basePath {
		return "", fmt.Errorf("path %q escapes storage root", objectName)
	}
	return abs, nil
}

func (s *LocalStorage) Upload(_ context.Context, objectName string, reader io.Reader, _ int64, _ string) error {
	dest, err := s.safePath(objectName)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("create parent dirs: %w", err)
	}

	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return f.Close()
}

func (s *LocalStorage) Download(_ context.Context, objectName string) (io.ReadCloser, error) {
	src, err := s.safePath(objectName)
	if err != nil {
		return nil, err
	}
	return os.Open(src)
}

func (s *LocalStorage) Delete(_ context.Context, objectName string) error {
	target, err := s.safePath(objectName)
	if err != nil {
		return err
	}
	if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *LocalStorage) HealthCheck(_ context.Context) error {
	info, err := os.Stat(s.basePath)
	if err != nil {
		return fmt.Errorf("storage dir: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("storage path %q is not a directory", s.basePath)
	}

	// Verify we can write by creating and immediately removing a temp file.
	tmp := filepath.Join(s.basePath, ".health-check")
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("storage not writable: %w", err)
	}
	f.Close()
	os.Remove(tmp)
	return nil
}
