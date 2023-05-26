package storage

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseVariantOptions(t *testing.T) {
	t.Parallel()

	params := make(url.Values)
	params.Add("size", "100")
	params.Add("resize_to_fill", "10x20")
	params.Add("format", "png")
	params.Add("quality", "85")
	params.Add("custom", "foo")
	options, err := ParseVariantOptions(params)
	require.NoError(t, err)

	require.Equal(t, 100, options.Size())
	resizeToFill, _ := options.ResizeToFill()
	require.Equal(t, [2]int{10, 20}, resizeToFill)
	require.Equal(t, "png", options.Format())
	require.Equal(t, 85, options.Quality())
	require.Equal(t, []string{"foo"}, options.Get("custom"))
}
