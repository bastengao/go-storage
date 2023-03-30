# storage

[![Go Reference](https://pkg.go.dev/badge/github.com/bastengao/go-storage.svg)](https://pkg.go.dev/github.com/bastengao/go-storage)

Inspired by ActiveStorage from Rails.

## Service

* [x] Disk
* [x] AWS S3
* [ ] Google Cloud Storage
* [ ] MicroSoft Azure Storage

## TODO

* [x] Public URL
* [ ] Private URL

## Documentation

* [Service AWS S3](service_s3.md)

## Usage

Use `Service` to manipulate files.

```go
import (
  "bytes"
  "context"
  "log"

  "github.com/aws/aws-sdk-go-v2/config"
  "github.com/bastengao/go-storage"
)

cfg, err := config.LoadDefaultConfig(context.TODO())
if err != nil {
  log.Fatal(err)
}

service, err := storage.NewS3(cfg, "bucket", "https://bucket.us-east-1.s3.amazonaws.com")
if err != nil {
  log.Fatal(err)
}

err = service.Upload(context.TODO(), "test/abc.txt", bytes.NewReader([]byte("hello world")))
if err != nil {
  log.Fatal(err)
}

err = service.Delete(context.TODO(), "test/abc.txt")
if err != nil {
  log.Fatal(err)
}
```

### Transforming Images

```go
store = storage.New(service, nil)

// transform image "sample.jpg" to a new jpeg with quality 75 and size 100x100 
options := storage.VariantOptions{}.
  SetFormat("jpeg").
  SetSize(100).
  SetQuality(75)
variant := store.Variant("sample.jpg", options)
err = variant.Process() // generate variant automatically if variant not exits
if err != nil {
  log.Fatal(err)
}
key := variant.Key()
// variants/sample-9c1aec99c27e6c66d895519f9ef831da.jpeg
```

### Serve files and variants

`server.Handler` will redirect to the actual service endpoint. This indirection decouples the service URL from the actual one.

```go
server := storage.NewServer("http://127.0.0.1:8080/storage/redirect", store, nil, nil)
url := server.URL("sample.jpg", nil)
// http://127.0.0.1:8080/storage/redirect?key=sample.jpg
variantURL := server.URL("sample.jpg", options)
// http://127.0.0.1:8080/storage/redirect?key=sample.jpg&format=jpeg&size=100&quality=75
// accessing variantURL will generate variant automatically if variant not exits

http.Handle("/storage/redirect", server.Handler())
err = http.ListenAndServe(":8080", nil)
if err != nil {
  panic(err)
}
```
