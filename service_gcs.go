package storage

import (
	"context"
	"errors"
	"io"

	gstorage "cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

const (
	Ctx_GCS_ACL contextKey = "gcs_acl"
)

type acl struct {
	entity gstorage.ACLEntity
	role   gstorage.ACLRole
}

func WithGcsACL(ctx context.Context, entity gstorage.ACLEntity, role gstorage.ACLRole) context.Context {
	list := gcsACLFromContext(ctx)
	list = append(list, acl{
		entity: entity,
		role:   role,
	})

	return context.WithValue(ctx, Ctx_GCS_ACL, list)
}

func WithGcsAllUsersRead(ctx context.Context) context.Context {
	list := gcsACLFromContext(ctx)
	list = append(list, acl{
		entity: gstorage.AllUsers,
		role:   gstorage.RoleReader,
	})

	return context.WithValue(ctx, Ctx_GCS_ACL, list)
}

func gcsACLFromContext(ctx context.Context) []acl {
	if v, ok := ctx.Value(Ctx_GCS_ACL).([]acl); ok {
		return v
	}

	return nil
}

type gcsService struct {
	client   *gstorage.Client
	bucket   string
	endpoint string
}

func NewGCSService(bucket string, endpoint string) (Service, error) {
	client, err := gstorage.NewClient(context.Background())
	if err != nil {
		return nil, err
	}

	return &gcsService{
		client:   client,
		bucket:   bucket,
		endpoint: endpoint,
	}, nil
}

func NewGCSServiceWithClient(bucket string, endpoint string, client *gstorage.Client) Service {
	return &gcsService{
		client:   client,
		bucket:   bucket,
		endpoint: endpoint,
	}
}

func (s *gcsService) Upload(ctx context.Context, key string, reader io.Reader) error {
	bucket := s.client.Bucket(s.bucket)
	obj := bucket.Object(key)

	writer := obj.NewWriter(ctx)
	_, err := io.Copy(writer, reader)
	if err != nil {
		writer.Close()
		return err
	}
	// NOTE: must close writer, otherwise the object will be not found before set ACL
	writer.Close()

	acl := gcsACLFromContext(ctx)
	for _, rule := range acl {
		err = obj.ACL().Set(ctx, rule.entity, rule.role)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *gcsService) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	bucket := s.client.Bucket(s.bucket)
	obj := bucket.Object(key)
	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (s *gcsService) Copy(ctx context.Context, src string, dst string) error {
	bucket := s.client.Bucket(s.bucket)
	srcObj := bucket.Object(src)
	dstObj := bucket.Object(dst)

	copier := dstObj.CopierFrom(srcObj)
	_, err := copier.Run(ctx)
	return err
}

func (s *gcsService) Delete(ctx context.Context, key string) error {
	bucket := s.client.Bucket(s.bucket)
	obj := bucket.Object(key)
	return obj.Delete(ctx)
}

func (s *gcsService) DeleteBatch(ctx context.Context, keys []string) error {
	// NOTE: SDK does not support batch delete
	for _, key := range keys {
		err := s.Delete(ctx, key)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *gcsService) DeletePrefixed(ctx context.Context, prefix string) error {
	bucket := s.client.Bucket(s.bucket)
	iter := bucket.Objects(ctx, &gstorage.Query{
		Prefix: prefix,
	})

	for {
		attrs, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return err
		}

		obj := bucket.Object(attrs.Name)
		err = obj.Delete(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *gcsService) Exist(ctx context.Context, key string) (bool, error) {
	bucket := s.client.Bucket(s.bucket)
	obj := bucket.Object(key)
	_, err := obj.Attrs(ctx)
	if err != nil {
		if errors.Is(err, gstorage.ErrObjectNotExist) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *gcsService) URL(key string) string {
	return URL(s.endpoint, key)
}
