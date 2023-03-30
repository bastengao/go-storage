package storage

import (
	"net/url"
	"strconv"
)

type VariantOptions map[string]any

func ParseVariantOptions(query url.Values) (VariantOptions, error) {
	options := make(VariantOptions)

	for k, v := range query {
		switch k {
		case "size":
			size, err := strconv.Atoi(v[0])
			if err != nil {
				return nil, err
			}
			options.SetSize(size)
		}
		// TODO
	}

	return options, nil
}

func (o VariantOptions) Size() int {
	s, _ := o["size"].(int)
	return s
}

// SetSize sets the size of the variant.
// Cut the image to a square from the center and resize it to the given size.
func (o VariantOptions) SetSize(size int) VariantOptions {
	o["size"] = size
	return o
}

func (o VariantOptions) ResizeToFill() ([2]int, bool) {
	v, ok := o["resize_to_fill"].([2]int)
	return v, ok
}

// SetResizeToFill sets the with and height of the variant.
// Cut the image to a rectangle from the center and resize it to the given size.
func (o VariantOptions) SetResizeToFill(size [2]int) VariantOptions {
	o["resize_to_fill"] = size
	return o
}

func (o VariantOptions) Format() string {
	f, _ := o["format"].(string)
	return f
}

// SetFormat sets the format of the variant. Must be one of "jpg", "png".
func (o VariantOptions) SetFormat(format string) VariantOptions {
	o["format"] = format
	return o
}

func (o VariantOptions) Quality() int {
	q, _ := o["quality"].(int)
	return q
}

// SetQuality sets the quality of the variant. Must be between 1 and 100. Default is 80.
func (o VariantOptions) SetQuality(quality int) VariantOptions {
	o["quality"] = quality
	return o
}

func (o VariantOptions) URLQuery() map[string]string {
	if len(o) == 0 {
		return nil
	}

	query := make(map[string]string)
	for k := range o {
		switch k {
		case "size":
			query[k] = strconv.Itoa(o.Size())
		case "resize_to_fill":
			size, ok := o.ResizeToFill()
			if ok {
				query[k] = strconv.Itoa(size[0]) + "x" + strconv.Itoa(size[1])
			}
		case "format":
			query[k] = o.Format()
		case "quality":
			query[k] = strconv.Itoa(o.Quality())
		}
	}
	return query
}
