package storage

import (
	"mime"
	"net/url"
	"path"
	"strings"
)

func URL(endpoint, p string) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		return ""
	}
	u.Path = path.Join(u.Path, p)
	return u.String()
}

var customMimeTypes = map[string]string{
	".heic": "image/heic",
}

func MimeTypeByExtension(ext string) string {
	m := mime.TypeByExtension(ext)
	if m != "" {
		return m
	}

	m, ok := customMimeTypes[strings.ToLower(ext)]
	if ok {
		return m
	}

	return ""
}
