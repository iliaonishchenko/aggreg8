package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type Signature struct {
	key string
}

func NewSignature(key string) *Signature {
	return &Signature{key: key}
}

func (s *Signature) Sign(src []byte) string {
	var dstStr string
	h := hmac.New(sha256.New, []byte(s.key))
	h.Write(src)
	dst := h.Sum(nil)
	dstStr = hex.EncodeToString(dst)
	return dstStr
}

func (s *Signature) Key() string {
	return s.key
}
