package storage

import (
	"context"
	"io"
)

type Service interface {
	Upload(ctx context.Context, key string, reader io.Reader) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Copy(ctx context.Context, src string, dst string) error
	Delete(ctx context.Context, key string) error
	DeleteBatch(keys []string) error
	DeletePrefixed(prefix string) error
	Exist(key string) (bool, error)
	URL(key string) string
}
