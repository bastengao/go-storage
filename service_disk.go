package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	pkgerr "github.com/pkg/errors"
)

var _ Service = (*disk)(nil)

type disk struct {
	dir      string
	endpoint string
}

// NewDiskService creates a new disk service.
// dir is the directory to store the files.
func NewDiskService(dir string, endpoint string) (Service, error) {
	_, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	return &disk{
		dir:      dir,
		endpoint: endpoint,
	}, nil
}

func (d *disk) Upload(ctx context.Context, key string, reader io.Reader) error {
	p, err := d.makePathFor(key)
	if err != nil {
		return err
	}

	out, err := os.Create(p)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, reader)
	return err
}

func (d *disk) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	f, err := os.Open(d.pathFor(key))
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (d *disk) Copy(ctx context.Context, src string, dst string) error {
	f, err := os.Open(d.pathFor(src))
	if err != nil {
		return err
	}
	defer f.Close()

	p, err := d.makePathFor(dst)
	if err != nil {
		return err
	}

	out, err := os.Create(p)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, f)
	return err
}

func (d *disk) Delete(ctx context.Context, key string) error {
	p := d.pathFor(key)
	err := os.Remove(p)
	if err != nil && !os.IsNotExist(err) {
		return pkgerr.WithStack(err)
	}
	return nil
}

func (d *disk) DeleteBatch(ctx context.Context, keys []string) error {
	for _, key := range keys {
		err := d.Delete(context.TODO(), key)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *disk) DeletePrefixed(ctx context.Context, prefix string) error {
	root := os.DirFS(d.dir)
	matches, err := fs.Glob(root, fmt.Sprintf("%s*", prefix))
	if err != nil {
		return pkgerr.WithStack(err)
	}

	for _, match := range matches {
		p := filepath.Join(d.dir, match)
		err := os.Remove(p)
		if err != nil {
			return pkgerr.WithStack(err)
		}
	}

	return nil
}

func (d *disk) Exist(ctx context.Context, key string) (bool, error) {
	_, err := os.Stat(d.pathFor(key))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (d *disk) URL(key string) string {
	return URL(d.endpoint, key)
}

func (d *disk) SignURL(ctx context.Context, key string, method string, expiresIn time.Duration) (string, http.Header, error) {
	return "", nil, errors.New("not supported")
}

func (d *disk) pathFor(key string) string {
	return filepath.Join(d.dir, key)
}

func (d *disk) makePathFor(key string) (string, error) {
	p := d.pathFor(key)
	err := os.MkdirAll(filepath.Dir(p), 0750)
	if err != nil && !os.IsExist(err) {
		return "", err
	}
	return p, nil
}
