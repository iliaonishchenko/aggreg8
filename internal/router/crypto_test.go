package router

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	cryptopkg "github.com/iliaonishchenko/aggreg8/internal/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithDecryption_DecryptsBody(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	enc, err := cryptopkg.NewEncrypterFromKey(&priv.PublicKey)
	require.NoError(t, err)
	dec, err := cryptopkg.NewDecrypterFromKey(priv)
	require.NoError(t, err)

	plaintext := []byte("hello world")
	encBody, encKey, err := enc.Encrypt(plaintext)
	require.NoError(t, err)

	var seen []byte
	handler := WithDecryption(dec)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/", bytes.NewReader(encBody))
	req.Header.Set("X-Crypto-Key", base64.StdEncoding.EncodeToString(encKey))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, plaintext, seen)
}

func TestWithDecryption_NoHeaderPassesThrough(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	dec, err := cryptopkg.NewDecrypterFromKey(priv)
	require.NoError(t, err)

	called := false
	handler := WithDecryption(dec)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, []byte("plain"), body)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte("plain")))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWithDecryption_BadHeaderReturns400(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	dec, err := cryptopkg.NewDecrypterFromKey(priv)
	require.NoError(t, err)

	handler := WithDecryption(dec)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte("garbage")))
	req.Header.Set("X-Crypto-Key", "!!!not-base64!!!")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWithDecryption_BadCiphertextReturns400(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	enc, err := cryptopkg.NewEncrypterFromKey(&priv.PublicKey)
	require.NoError(t, err)
	dec, err := cryptopkg.NewDecrypterFromKey(priv)
	require.NoError(t, err)

	_, encKey, err := enc.Encrypt([]byte("hi"))
	require.NoError(t, err)

	handler := WithDecryption(dec)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte("not-real-ciphertext")))
	req.Header.Set("X-Crypto-Key", base64.StdEncoding.EncodeToString(encKey))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
