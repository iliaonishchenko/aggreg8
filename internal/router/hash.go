package router

import (
	"bytes"
	"io"
	"net/http"
)

type Signature interface {
	Sign(data []byte) string
	Key() string
}

func WithHash(signature Signature) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if signature.Key() == "" {
				h.ServeHTTP(w, r)
				return
			}
			bodyBytes, err := io.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				http.Error(w, "failed to ready body", http.StatusInternalServerError)
				return
			}
			hHash := r.Header.Get("HashSHA256")
			if hHash != "" {
				rHash := signature.Sign(bodyBytes)
				if hHash != rHash {
					http.Error(w, "invalid hash", http.StatusBadRequest)
					return
				}
			}
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			wrapper := &responseWrapper{ResponseWriter: w, body: &bytes.Buffer{}}

			h.ServeHTTP(wrapper, r)

			responseHash := signature.Sign(wrapper.body.Bytes())
			w.Header().Set("HashSHA256", responseHash)
			w.Write(wrapper.body.Bytes())
		})
	}
}

type responseWrapper struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWrapper) Write(b []byte) (int, error) {
	return w.body.Write(b)
}
