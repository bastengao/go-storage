package storage

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"
)

var _ Service = (*NullService)(nil)

type NullService struct{}

func NewNullService() Service {
	return NullService{}
}

func (NullService) Upload(ctx context.Context, key string, reader io.Reader) error {
	return nil
}

func (NullService) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, nil
}

func (NullService) Copy(ctx context.Context, src string, dst string) error {
	return nil
}

func (NullService) Delete(ctx context.Context, key string) error {
	return nil
}

func (NullService) DeleteBatch(ctx context.Context, keys []string) error {
	return nil
}

func (NullService) DeletePrefixed(ctx context.Context, prefix string) error {
	return nil
}

func (NullService) Exist(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (NullService) URL(key string) string {
	return ""
}

func (NullService) SignURL(ctx context.Context, key string, method string, expiresIn time.Duration) (string, http.Header, error) {
	return "", nil, errors.New("not supported")
}
