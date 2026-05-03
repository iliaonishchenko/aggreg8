package router

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
)

type Decrypter interface {
	Decrypt(body []byte, encKey []byte) ([]byte, error)
}

func WithDecryption(d Decrypter) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			keyHeader := r.Header.Get("X-Crypto-Key")
			if keyHeader == "" {
				h.ServeHTTP(w, r)
				return
			}
			encKey, err := base64.StdEncoding.DecodeString(keyHeader)
			if err != nil {
				http.Error(w, "неверный X-Crypto-Key хедер", http.StatusBadRequest)
				return
			}
			body, err := io.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				http.Error(w, "ошибка чтения тела запроса", http.StatusInternalServerError)
				return
			}
			plaintext, err := d.Decrypt(body, encKey)
			if err != nil {
				http.Error(w, "ошибка дешифровки тела запроса", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(plaintext))
			r.ContentLength = int64(len(plaintext))
			h.ServeHTTP(w, r)
		})
	}
}
