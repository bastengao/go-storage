# Service S3

## Initial

```go
import (
  "github.com/aws/aws-sdk-go-v2/config"
  "github.com/bastengao/go-storage"
)

// load config from env, see details https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/
cfg, err := config.LoadDefaultConfig(context.TODO())
if err != nil {
  log.Fatal(err)
}

service, err := storage.NewS3(
  cfg,
  "bucket",
  "https://bucket.us-east-1.s3.amazonaws.com",
)
```

## Default ACL

```go
import (
  "github.com/aws/aws-sdk-go-v2/service/s3/types"
  "github.com/bastengao/go-storage"
)

defaultACL := types.ObjectCannedACLPublicRead
service, err := storage.NewS3(cfg, bucket, endpoint, storage.S3Options{
    ACL: &defaultACL,
})
```

## Custom ACL

```go
ctx := storage.WithS3Private(context.TODO())
err = service.Upload(ctx, key, reader)
```

## Specify content-type

```go
ctx := storage.WithS3ContentType(ctx, "text/plain")
err = service.Upload(ctx, key, reader)
```
