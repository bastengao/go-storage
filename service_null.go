package storage

import (
	"context"
	"io"
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
	// TODO
	return nil, nil
}

func (NullService) Copy(ctx context.Context, src string, dst string) error {
	return nil
}

func (NullService) Delete(ctx context.Context, prefix string) error {
	return nil
}

func (NullService) DeleteBatch(keys []string) error {
	return nil
}

func (NullService) DeletePrefixed(key string) error {
	return nil
}

func (NullService) Exist(key string) (bool, error) {
	return false, nil
}

func (NullService) URL(key string) string {
	return ""
}
