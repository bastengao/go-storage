package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestURLSigner(t *testing.T) {
	t.Parallel()

	t.Run("Sign", func(t *testing.T) {
		signer := NewHmacURLSigner([]byte("key"))
		signedURL, err := signer.Sign("http://example.com?foo=bar", time.Hour)
		require.NoError(t, err)
		t.Log(signedURL)
	})

	t.Run("Sign without query", func(t *testing.T) {
		signer := NewHmacURLSigner([]byte("key"))
		signedURL, err := signer.Sign("http://example.com", 0)
		require.NoError(t, err)
		t.Log(signedURL)
	})

	t.Run("Validate", func(t *testing.T) {
		signer := NewHmacURLSigner([]byte("key"))
		signedURL, err := signer.Sign("http://example.com?foo=bar", time.Hour)
		require.NoError(t, err)

		err = signer.Validate(signedURL)
		require.NoError(t, err)
	})
}
