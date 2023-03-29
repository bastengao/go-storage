package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedirectURL(t *testing.T) {
	options := VariantOptions{}.
		SetFormat("jpeg").
		SetSize(100).
		SetQuality(75)
	url := RedirectURL("http://example.com", "images/test.png", options)
	require.Equal(t, "http://example.com?format=jpeg&key=images%2Ftest.png&quality=75&size=100", url)
}
