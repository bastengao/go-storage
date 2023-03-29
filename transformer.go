package storage

import (
	"context"
	"image"
	"io"
	"math"

	"github.com/disintegration/imaging"
	pkgerr "github.com/pkg/errors"
)

type Transformer interface {
	Transform(ctx context.Context, options VariantOptions, format string, source io.Reader, writer io.Writer) error
}

type defaultTransformer struct{}

func NewTransformer() Transformer {
	return defaultTransformer{}
}

func (defaultTransformer) Transform(ctx context.Context, options VariantOptions, format string, source io.Reader, writer io.Writer) error {
	img, err := DecodeImg(source)
	if err != nil {
		return pkgerr.WithStack(err)
	}

	size := options.Size()
	if size != 0 {
		img = CropResize(img, int(size))
	}

	dimension, ok := options.ResizeToFill()
	if ok {
		width := dimension[0]
		height := dimension[1]
		img = CropResize2(img, width, height)
	}

	// TODO: support webp
	err = imaging.Encode(writer, img, imagingFormat(format), encodeOptions(options, format)...)
	if err != nil {
		return pkgerr.WithStack(err)
	}

	return nil
}

func DecodeImg(r io.Reader) (image.Image, error) {
	return imaging.Decode(r, imaging.AutoOrientation(true))
}

// CropCenter crop center
func CropCenter(img image.Image) image.Image {
	width, height := imageDim(img)

	// crop to square
	minSize := int(math.Min(float64(width), float64(height)))
	return imaging.CropCenter(img, minSize, minSize)
}

// CropResize crop center and resize to square
func CropResize(img image.Image, size int) image.Image {
	out := CropCenter(img)

	// resize
	return imaging.Resize(out, size, size, imaging.Lanczos)
}

// CropResize2 crop center and resize to specified size
func CropResize2(img image.Image, width int, height int) image.Image {
	originWidth, originHeight := imageDim(img)

	widthRatio := float64(originWidth) / float64(width)
	heightRatio := float64(originHeight) / float64(height)
	minRatio := math.Min(widthRatio, heightRatio)

	var resizeW, resizeH int
	if minRatio == widthRatio {
		resizeW = width
		resizeH = int(float64(originHeight) / minRatio)
	} else if minRatio == heightRatio {
		resizeW = int(float64(originWidth) / minRatio)
		resizeH = height
	}
	// resize
	out := imaging.Resize(img, resizeW, resizeH, imaging.Lanczos)

	// crop
	return imaging.CropCenter(out, width, height)
}

func imageDim(img image.Image) (int, int) {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	return width, height
}

func imagingFormat(format string) imaging.Format {
	if format == "jpeg" {
		return imaging.JPEG
	} else if format == "png" {
		return imaging.PNG
	}
	return imaging.JPEG
}

func encodeOptions(options VariantOptions, format string) []imaging.EncodeOption {
	if imagingFormat(format) == imaging.JPEG {
		if q := options.Quality(); q > 0 && q <= 90 {
			return []imaging.EncodeOption{imaging.JPEGQuality(q)}
		}
		return []imaging.EncodeOption{imaging.JPEGQuality(80)}
	}

	return nil
}
