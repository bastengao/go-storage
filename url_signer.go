package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type URLSigner interface {
	// exp will be ignored if exp is 0.
	Sign(url string, exp time.Duration) (string, error)
	// return error if invalid
	Validate(signedUrl string) error
}

type hmacUrlSigner struct {
	key []byte
}

func NewHmacURLSigner(key []byte) URLSigner {
	return &hmacUrlSigner{
		key: key,
	}
}

func (s hmacUrlSigner) Sign(plainUrl string, exp time.Duration) (string, error) {
	u, err := url.Parse(plainUrl)
	if err != nil {
		return "", err
	}

	query := u.Query()

	if exp != 0 {
		expires := strconv.FormatInt(time.Now().Add(exp).Unix(), 10)
		query.Add("expires", expires)
	}

	u.RawQuery = query.Encode()
	readyToSign := u.String()
	signature := hashString(s.key, readyToSign)

	query.Add("signature", signature)
	u.RawQuery = query.Encode()
	signedURL := u.String()

	return signedURL, nil
}

func (s hmacUrlSigner) Validate(signedUrl string) error {
	u, err := url.Parse(signedUrl)
	if err != nil {
		return err
	}

	query := u.Query()

	if query.Has("expires") {
		expires := query.Get("expires")
		unix, err := strconv.ParseInt(expires, 10, 64)
		if err != nil {
			return errors.Wrap(err, "invalid expires")
		}
		t := time.Unix(unix, 0)
		if t.Before(time.Now()) {
			return errors.New("expired")
		}
	}

	if !query.Has("signature") {
		return errors.New("missing signature")
	}

	signature := query.Get("signature")
	query.Del("signature")
	u.RawQuery = query.Encode()
	computedSignature := hashString(s.key, u.String())

	if computedSignature != signature {
		return errors.New("invalid signature")
	}

	return nil
}

func hashString(key []byte, s string) string {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
