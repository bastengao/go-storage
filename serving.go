package storage

import (
	"net/http"
	"net/url"
)

type Server interface {
	Handler() http.Handler
	URL(key string, options VariantOptions) string
}

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

type storageServer struct {
	endpoint   string
	storage    Storage
	keyEncoder func(string) string
	keyDecoder func(string) string
}

// NewServer creates a new server. keyEncoder and keyDecoder are optional.
//
// Default key will keep unchanged in query, such as "key=sample.jpg". keyEncoder and keyDecoder can be used to encode/decode key.
func NewServer(endpoint string, storage Storage, keyEncoder func(string) string, keyDecoder func(string) string) Server {
	return storageServer{
		endpoint:   endpoint,
		storage:    storage,
		keyEncoder: keyEncoder,
		keyDecoder: keyDecoder,
	}
}

func (s storageServer) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		options, err := ParseVariantOptions(r.URL.Query())
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
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
			url := s.storage.Service().URL(key)
			http.Redirect(w, r, url, http.StatusFound)
			return
		}

		// variant file
		variant := s.storage.Variant(key, options)
		err = variant.Process()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, variant.URL(), http.StatusFound)
	})
}

// URL returns the URL of the variant serving by this server.
func (s storageServer) URL(key string, options VariantOptions) string {
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
	return u.String()
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
