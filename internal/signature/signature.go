// Package signature реализует HMAC-SHA256 подпись данных.
package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Signature реализует HMAC-SHA256 подпись с заданным секретным ключом.
type Signature struct {
	key string
}

// NewSignature создаёт новый Signature с указанным секретным ключом.
func NewSignature(key string) *Signature {
	return &Signature{key: key}
}

// Sign вычисляет HMAC-SHA256 подпись и возвращает её в hex-формате.
func (s *Signature) Sign(src []byte) string {
	var dstStr string
	h := hmac.New(sha256.New, []byte(s.key))
	h.Write(src)
	dst := h.Sum(nil)
	dstStr = hex.EncodeToString(dst)
	return dstStr
}

// Key возвращает секретный ключ подписи.
func (s *Signature) Key() string {
	return s.key
}
