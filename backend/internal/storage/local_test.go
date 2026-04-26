package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func tempStorage(t *testing.T) *LocalStorage {
	t.Helper()
	dir := t.TempDir()
	s, err := NewLocalStorage(dir)
	if err != nil {
		t.Fatalf("NewLocalStorage: %v", err)
	}
	return s
}

func TestLocalStorage_UploadDownloadDelete(t *testing.T) {
	s := tempStorage(t)
	ctx := context.Background()
	data := []byte("hello, world")

	if err := s.Upload(ctx, "test/file.txt", bytes.NewReader(data), int64(len(data)), "text/plain"); err != nil {
		t.Fatalf("Upload: %v", err)
	}

	rc, err := s.Download(ctx, "test/file.txt")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	got, _ := io.ReadAll(rc)
	rc.Close()

	if !bytes.Equal(got, data) {
		t.Fatalf("got %q, want %q", got, data)
	}

	if err := s.Delete(ctx, "test/file.txt"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := s.Download(ctx, "test/file.txt"); err == nil {
		t.Fatal("expected error downloading deleted file")
	}
}

func TestLocalStorage_DeleteNonExistent(t *testing.T) {
	s := tempStorage(t)
	if err := s.Delete(context.Background(), "nope.txt"); err != nil {
		t.Fatalf("Delete non-existent: %v", err)
	}
}

func TestLocalStorage_PathTraversal(t *testing.T) {
	s := tempStorage(t)
	ctx := context.Background()

	cases := []string{
		"../etc/passwd",
		"../../etc/shadow",
		"foo/../../etc/passwd",
		"/etc/passwd",
	}

	for _, name := range cases {
		if err := s.Upload(ctx, name, bytes.NewReader([]byte("x")), 1, ""); err == nil {
			t.Errorf("Upload(%q) should have failed", name)
		}
		if _, err := s.Download(ctx, name); err == nil {
			t.Errorf("Download(%q) should have failed", name)
		}
		if err := s.Delete(ctx, name); err == nil {
			t.Errorf("Delete(%q) should have failed", name)
		}
	}
}

func TestLocalStorage_HealthCheck(t *testing.T) {
	s := tempStorage(t)
	if err := s.HealthCheck(context.Background()); err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
}

func TestLocalStorage_HealthCheckBadDir(t *testing.T) {
	s := &LocalStorage{basePath: filepath.Join(os.TempDir(), "nonexistent-moneyapp-test-dir-xyz")}
	if err := s.HealthCheck(context.Background()); err == nil {
		t.Fatal("expected error for missing dir")
	}
}

func TestLocalStorage_ImplementsObjectStore(t *testing.T) {
	var _ ObjectStore = (*LocalStorage)(nil)
}

func TestMinIOStorage_ImplementsObjectStore(t *testing.T) {
	var _ ObjectStore = (*MinIOStorage)(nil)
}
