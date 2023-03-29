package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/bastengao/go-storage"
)

func exampleS3() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	defaultACL := types.ObjectCannedACLPublicRead
	service, err := storage.NewS3(
		cfg,
		"bucket",
		"https://bucket.us-east-1.s3.amazonaws.com",
		storage.S3Options{
			ACL: &defaultACL,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// custom ACL
	ctx := storage.WithS3Private(context.TODO())
	err = service.Upload(ctx, "test/abc.txt", bytes.NewReader([]byte("hello world")))
	if err != nil {
		log.Fatal(err)
	}
	err = service.Delete(context.TODO(), "test/abc.txt")
	if err != nil {
		log.Fatal(err)
	}

	store = storage.New(service, nil)

	// transform image to 100x100 jpeg with quality 75
	options := storage.VariantOptions{}.
		SetFormat("jpeg").
		SetSize(100).
		SetQuality(75)
	variant := store.Variant("sample.jpg", options)
	err = variant.Process()
	if err != nil {
		log.Fatal(err)
	}
	key := variant.Key()
	// variants/sample-9c1aec99c27e6c66d895519f9ef831da.jpeg
	_ = key

	server := storage.NewServer("http://127.0.0.1:8080/storage/redirect", store, nil, nil)
	url := server.URL("sample.jpg", nil)
	// http://127.0.0.1:8080/storage/redirect?key=sample.jpg
	variantURL := server.URL("sample.jpg", options)
	// http://127.0.0.1:8080/storage/redirect?key=sample.jpg&format=jpeg&size=100&quality=75
	_ = url
	_ = variantURL

	http.Handle("/storage/redirect", server.Handler())
	fmt.Println("Listening on http://127.0.0.1:8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
