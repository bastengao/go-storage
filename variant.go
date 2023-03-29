package storage

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/disintegration/imaging"
)

type Variant interface {
	Process() error
	Key() string
	URL() string
}

type variant struct {
	service     Service
	originKey   string
	options     VariantOptions
	transformer Transformer
}

func NewVariant(service Service, originKey string, options VariantOptions, transformer Transformer) Variant {
	return variant{
		service:     service,
		originKey:   originKey,
		options:     options,
		transformer: transformer,
	}
}

func (v variant) Process() error {
	// skip if already exists
	exist, err := v.service.Exist(v.Key())
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	reader, err := v.service.Download(context.TODO(), v.originKey)
	if err != nil {
		return err
	}
	defer reader.Close()

	ctx := context.TODO()
	var buff bytes.Buffer
	err = v.transformer.Transform(ctx, v.options, v.format(), reader, &buff)
	if err != nil {
		return err
	}

	// TODO: set S3 ACL ctx := WithS3PublicRead(context.TODO())
	err = v.service.Upload(ctx, v.Key(), &buff)
	if err != nil {
		return err
	}
	return nil
}

func (v variant) URL() string {
	return v.service.URL(v.Key())
}

func (v variant) Key() string {
	dir, file := path.Split(v.originKey)
	originExt := path.Ext(v.originKey)
	format := v.format()
	ext := "." + strings.ToLower(format)
	basename := strings.TrimSuffix(file, originExt) + "-" + v.digest() + ext
	return path.Join("variants", dir, basename)
}

func (v variant) format() string {
	if f := v.options.Format(); f != "" {
		switch f {
		case "jpeg", "png", "webp":
			return f
		default:
			return "jpeg"
		}
	}

	format, err := imaging.FormatFromExtension(path.Ext(v.originKey))
	if err != nil {
		if errors.Is(err, imaging.ErrUnsupportedFormat) {
			return "jpeg"
		}

		format = imaging.JPEG
	}

	return format.String()
}

func (v variant) digest() string {
	b, err := json.Marshal(v.options) // TODO: should sort keys
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%x", md5.Sum(b))
}
