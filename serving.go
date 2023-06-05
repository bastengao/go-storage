package storage

import (
	"net/http"
	"net/url"
	"time"
)

type URLOptions struct {
	expires *time.Duration
}

type URLOption func(o *URLOptions)

func (o URLOption) Apply(options *URLOptions) {
	o(options)
}

// WithURLExpires sets the expires of the serving URL. 0 means never expires.
func WithURLExpires(expires time.Duration) URLOption {
	return func(o *URLOptions) {
		o.expires = &expires
	}
}

type Server interface {
	Handler() http.Handler
	URL(key string, options VariantOptions, urlOptions ...URLOption) string
}

// RedirectURL returns serving URL of the variant.
// Deprecated: out of date.
func RedirectURL(endpoint string, key string, options VariantOptions) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		return ""
	}

	query := u.Query()
	query.Add("key", key)

	if options != nil {
		for k, v := range options.URLQuery() {
			query.Add(k, v)
		}
	}

	u.RawQuery = query.Encode()
	return u.String()
}

type ServerOptions struct {
	KeyEncoder func(key string) string
	KeyDecoder func(encodedKey string) string
	// URLResolver is used to resolve the URL of the variant or origin file.
	// Default will use Service.URL(key) method
	URLResolver func(key string) string
	// Will sign serving URL to prevent somebody to change serving URL
	SigningKey []byte
	// Expire duration for signed serving URL
	SigningExpires time.Duration
}

type ServerOption func(o *ServerOptions)

type storageServer struct {
	endpoint       string
	storage        Storage
	keyEncoder     func(string) string
	keyDecoder     func(string) string
	urlResolver    func(string) string
	urlSigner      URLSigner
	signingExpires time.Duration
}

// NewServer creates a new server. keyEncoder and keyDecoder are optional.
//
// Default key will keep unchanged in query, such as "key=sample.jpg". keyEncoder and keyDecoder can be used to encode/decode key.
func NewServer(endpoint string, storage Storage, options ...ServerOption) Server {
	opts := &ServerOptions{
		URLResolver: func(key string) string {
			return storage.Service().URL(key)
		},
	}
	for _, opt := range options {
		opt(opts)
	}

	var urlSigner URLSigner
	if opts.SigningKey != nil {
		urlSigner = NewHmacURLSigner(opts.SigningKey)
	}

	return storageServer{
		endpoint:       endpoint,
		storage:        storage,
		keyEncoder:     opts.KeyEncoder,
		keyDecoder:     opts.KeyDecoder,
		urlResolver:    opts.URLResolver,
		urlSigner:      urlSigner,
		signingExpires: opts.SigningExpires,
	}
}

func (s storageServer) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.urlSigner != nil {
			// NOTE: only path and RawQuery of URL are set
			fullURL := s.endpoint + "?" + r.URL.RawQuery
			err := s.urlSigner.Validate(fullURL)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
		}

		strippedQuery := r.URL.Query()
		strippedQuery.Del("key")
		strippedQuery.Del("signature")
		strippedQuery.Del("expires")
		options, err := ParseVariantOptions(strippedQuery)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		key := r.URL.Query().Get("key")
		if key == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		key = s.decodeKey(key)

		// origin file
		if len(options) == 0 {
			url := s.urlResolver(key)
			http.Redirect(w, r, url, http.StatusFound)
			return
		}

		// variant file
		variant := s.storage.Variant(key, options)
		err = variant.Process()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		url := s.urlResolver(variant.Key())
		http.Redirect(w, r, url, http.StatusFound)
	})
}

// URL returns the URL of the variant serving by this server.
func (s storageServer) URL(key string, options VariantOptions, urlOpts ...URLOption) string {
	var urlOptions URLOptions
	for _, opt := range urlOpts {
		opt.Apply(&urlOptions)
	}

	u, err := url.Parse(s.endpoint)
	if err != nil {
		return ""
	}

	query := u.Query()
	query.Add("key", s.encodeKey(key))

	if options != nil {
		for k, v := range options.URLQuery() {
			query.Add(k, v)
		}
	}

	u.RawQuery = query.Encode()
	if s.urlSigner == nil {
		return u.String()
	}

	expires := s.signingExpires
	if urlOptions.expires != nil {
		expires = *urlOptions.expires
	}
	signedURL, err := s.urlSigner.Sign(u.String(), expires)
	if err != nil {
		return ""
	}
	return signedURL
}

func (s storageServer) encodeKey(key string) string {
	if s.keyEncoder != nil {
		return s.keyEncoder(key)
	}
	return key
}

func (s storageServer) decodeKey(key string) string {
	if s.keyDecoder != nil {
		return s.keyDecoder(key)
	}
	return key
}
