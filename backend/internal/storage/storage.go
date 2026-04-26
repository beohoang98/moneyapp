package storage

import (
	"context"
	"io"
)

type ObjectStore interface {
	Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error
	Download(ctx context.Context, objectName string) (io.ReadCloser, error)
	Delete(ctx context.Context, objectName string) error
	HealthCheck(ctx context.Context) error
}
