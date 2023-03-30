package storage

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGCSService(t *testing.T) {
	t.Skip()

	service, err := NewGCSService("go-storage-test", "https://storage.cloud.google.com/go-storage-test")
	require.NoError(t, err)

	ok, err := service.Exist(context.TODO(), "test.txt")
	require.NoError(t, err)
	require.False(t, ok)

	err = service.Upload(context.TODO(), "test.txt", bytes.NewReader([]byte("hello world")))
	require.NoError(t, err)

	ok, err = service.Exist(context.TODO(), "test.txt")
	require.NoError(t, err)
	require.True(t, ok)

	err = service.Copy(context.TODO(), "test.txt", "test2.txt")
	require.NoError(t, err)

	reader, err := service.Download(context.TODO(), "test2.txt")
	require.NoError(t, err)
	defer reader.Close()
	b, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, "hello world", string(b))

	err = service.Delete(context.TODO(), "test.txt")
	require.NoError(t, err)
}

func TestGCSService_deletePrefix(t *testing.T) {
	t.Skip()

	service, err := NewGCSService("go-storage-test", "https://storage.cloud.google.com/go-storage-test")
	require.NoError(t, err)

	err = service.Upload(context.TODO(), "abc/test.txt", bytes.NewReader([]byte("hello world")))
	require.NoError(t, err)

	err = service.Upload(context.TODO(), "test.txt", bytes.NewReader([]byte("hello world")))
	require.NoError(t, err)

	err = service.Upload(context.TODO(), "test-1.txt", bytes.NewReader([]byte("hello world")))
	require.NoError(t, err)

	err = service.Upload(context.TODO(), "test-2.txt", bytes.NewReader([]byte("hello world")))
	require.NoError(t, err)

	err = service.DeletePrefixed(context.TODO(), "test")
	require.NoError(t, err)
}

func TestGCSService_uploadWithACL(t *testing.T) {
	t.Skip()

	service, err := NewGCSService("go-storage-test", "https://storage.cloud.google.com/go-storage-test")
	require.NoError(t, err)

	ctx := WithGcsAllUsersRead(context.TODO())
	err = service.Upload(ctx, "test.txt", bytes.NewReader([]byte("hello world")))
	require.NoError(t, err)
}
