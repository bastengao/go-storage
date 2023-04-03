package storage

import (
	"context"
	"io"
	"net/http"
	"time"
)

type Service interface {
	Upload(ctx context.Context, key string, reader io.Reader) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Copy(ctx context.Context, src string, dst string) error
	Delete(ctx context.Context, key string) error
	DeleteBatch(ctx context.Context, keys []string) error
	DeletePrefixed(ctx context.Context, prefix string) error
	Exist(ctx context.Context, key string) (bool, error)
	URL(key string) string
	// SignURL returns a signed URL for the given key.
	//
	// method must be one of "GET", "PUT", "HEAD", "DELETE".
	SignURL(ctx context.Context, key string, method string, expiresIn time.Duration) (string, http.Header, error)
}
