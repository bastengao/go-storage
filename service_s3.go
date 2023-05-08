package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	pkgerr "github.com/pkg/errors"
)

type contextKey string

const (
	CtxS3ACL         contextKey = "s3_acl"
	CtxS3ContentType contextKey = "s3_contentType"
)

// WithS3PublicRead set s3 object acl to public-read for upload and copy.
func WithS3PublicRead(ctx context.Context) context.Context {
	return context.WithValue(ctx, CtxS3ACL, types.ObjectCannedACLPublicRead)
}

// WithS3Private set s3 object acl to private for upload and copy.
func WithS3Private(ctx context.Context) context.Context {
	return context.WithValue(ctx, CtxS3ACL, types.ObjectCannedACLPrivate)
}

// WithS3ContentType set s3 object content-type for upload and copy.
func WithS3ContentType(ctx context.Context, contentType string) context.Context {
	return context.WithValue(ctx, CtxS3ContentType, contentType)
}

func s3ACLFromContext(ctx context.Context) *types.ObjectCannedACL {
	if v, ok := ctx.Value(CtxS3ACL).(types.ObjectCannedACL); ok {
		return &v
	}
	return nil
}

func contentTypeFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(CtxS3ContentType).(string); ok {
		return v
	}
	return ""
}

type S3Options struct {
	// config upload or copy ACL. Default is private
	ACL *types.ObjectCannedACL
}

var _ Service = (*s3Service)(nil)

type s3Service struct {
	svc        *s3.Client
	uploader   *manager.Uploader
	downloader *manager.Downloader
	bucket     string
	endpoint   string
	acl        types.ObjectCannedACL
}

func NewS3(cfg aws.Config, bucket string, endpoint string, options ...S3Options) (Service, error) {
	_, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	acl := types.ObjectCannedACLPrivate
	for _, opt := range options {
		if opt.ACL != nil {
			acl = *opt.ACL
		}
	}

	svc := s3.NewFromConfig(cfg)
	return &s3Service{
		svc:        svc,
		uploader:   manager.NewUploader(svc),
		downloader: manager.NewDownloader(svc),
		bucket:     bucket,
		endpoint:   endpoint,
		acl:        acl,
	}, nil
}

func (s *s3Service) Upload(ctx context.Context, key string, reader io.Reader) error {
	// detect content type from extension
	contentType := "application/octet-stream"
	if ct := MimeTypeByExtension(path.Ext(key)); ct != "" {
		contentType = ct
	}
	if ct := contentTypeFromContext(ctx); ct != "" {
		contentType = ct
	}
	// TODO: detect content type from content

	acl := s.acl
	if ctxACL := s3ACLFromContext(ctx); ctxACL != nil {
		acl = *ctxACL
	}

	_, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(s.bucket),
		Key:          aws.String(key),
		ACL:          acl,
		Body:         reader,
		ContentType:  aws.String(contentType),
		StorageClass: types.StorageClassIntelligentTiering,
	})
	return pkgerr.WithStack(err)
}

func (s *s3Service) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	var buf manager.WriteAtBuffer
	_, err := s.downloader.Download(
		ctx,
		&buf,
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		},
	)
	if err != nil {
		return nil, err
	}

	return manager.ReadSeekCloser(bytes.NewReader(buf.Bytes())), nil
}

func (s *s3Service) Copy(ctx context.Context, src string, dst string) error {
	// detect content type from extension
	contentType := "application/octet-stream"
	if ct := MimeTypeByExtension(path.Ext(dst)); ct != "" {
		contentType = ct
	}
	if ct := contentTypeFromContext(ctx); ct != "" {
		contentType = ct
	}

	acl := s.acl
	if ctxACL := s3ACLFromContext(ctx); ctxACL != nil {
		acl = *ctxACL
	}

	_, err := s.svc.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:            aws.String(s.bucket),
		Key:               aws.String(dst),
		ACL:               acl,
		MetadataDirective: types.MetadataDirectiveReplace,
		ContentType:       aws.String(contentType),
		StorageClass:      types.StorageClassIntelligentTiering,
		CopySource:        aws.String(fmt.Sprintf("%s/%s", s.bucket, src)),
	})
	return pkgerr.WithStack(err)
}

func (s *s3Service) Delete(ctx context.Context, key string) error {
	_, err := s.svc.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var ae smithy.APIError
		if ok := errors.As(err, &ae); ok {
			if ae.ErrorCode() == "NoSuchKey" || ae.ErrorCode() == "NotFound" {
				return nil
			}
		}
		return err
	}

	return nil
}

func (s *s3Service) DeleteBatch(ctx context.Context, keys []string) error {
	objects := make([]types.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objects[i] = types.ObjectIdentifier{Key: aws.String(key)}
	}
	_, err := s.svc.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(s.bucket),
		Delete: &types.Delete{
			Objects: objects,
		},
	})
	if err != nil {
		return pkgerr.WithStack(err)
	}

	return nil
}

func (s *s3Service) DeletePrefixed(ctx context.Context, prefix string) error {
	p := s3.NewListObjectsV2Paginator(s.svc, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})

	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return pkgerr.WithStack(err)
		}

		if len(page.Contents) == 0 {
			return nil
		}

		keys := make([]types.ObjectIdentifier, len(page.Contents))
		for i, obj := range page.Contents {
			keys[i] = types.ObjectIdentifier{Key: obj.Key}
		}
		_, err = s.svc.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(s.bucket),
			Delete: &types.Delete{
				Objects: keys,
			},
		})
		if err != nil {
			return pkgerr.WithStack(err)
		}
	}

	return nil
}

func (s *s3Service) Exist(ctx context.Context, key string) (bool, error) {
	_, err := s.svc.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		var ae smithy.APIError
		if ok := errors.As(err, &ae); ok {
			if ae.ErrorCode() == "NoSuchKey" || ae.ErrorCode() == "NotFound" {
				return false, nil
			}
		}

		return false, pkgerr.WithStack(err)
	}

	return true, nil
}

func (s *s3Service) URL(key string) string {
	return URL(s.endpoint, key)
}

func (s *s3Service) SignURL(ctx context.Context, key string, method string, expiresIn time.Duration) (string, http.Header, error) {
	presignedClient := s3.NewPresignClient(s.svc)
	var optFns []func(*s3.PresignOptions)
	if expiresIn != 0 {
		optFns = append(optFns, s3.WithPresignExpires(expiresIn))
	}

	switch method {
	case http.MethodGet:
		req, err := presignedClient.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		}, optFns...)
		if err != nil {
			return "", nil, err
		}
		return req.URL, req.SignedHeader, nil
	case http.MethodPut:
		req, err := presignedClient.PresignPutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		}, optFns...)
		if err != nil {
			return "", nil, err
		}
		return req.URL, req.SignedHeader, nil
	case http.MethodHead:
		req, err := presignedClient.PresignHeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		}, optFns...)
		if err != nil {
			return "", nil, err
		}
		return req.URL, req.SignedHeader, nil
	case http.MethodDelete:
		req, err := presignedClient.PresignDeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		}, optFns...)
		if err != nil {
			return "", nil, err
		}
		return req.URL, req.SignedHeader, nil
	}

	return "", nil, errors.New("not supported")
}
