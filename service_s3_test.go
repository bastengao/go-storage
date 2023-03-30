package storage

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/require"
)

func TestS3Upload(t *testing.T) {
	// TODO: use mock
	t.Skip()

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithClientLogMode(aws.LogRetries|aws.LogRequest|aws.LogResponse))
	require.NoError(t, err)

	s3, err := NewS3(cfg, "TODO: bucket", "TODO: endpoint")
	require.NoError(t, err)

	ctx := WithS3PublicRead(context.TODO())
	err = s3.Upload(ctx, "test/abc.txt", bytes.NewReader([]byte("hello world")))
	require.NoError(t, err)
}

func TestS3ACLFromContext(t *testing.T) {
	acl := s3ACLFromContext(context.TODO())
	require.Nil(t, acl)

	ctx := WithS3PublicRead(context.TODO())
	acl = s3ACLFromContext(ctx)
	require.Equal(t, types.ObjectCannedACLPublicRead, *acl)
}
