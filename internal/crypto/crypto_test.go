package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writePEM(t *testing.T, dir, name, blockType string, der []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()
	require.NoError(t, pem.Encode(f, &pem.Block{Type: blockType, Bytes: der}))
	return path
}

func TestLoadPublicKey_PKIX(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	der, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	require.NoError(t, err)
	path := writePEM(t, t.TempDir(), "pub.pem", "PUBLIC KEY", der)

	enc, err := LoadPublicKey(path)
	require.NoError(t, err)
	assert.NotNil(t, enc)
}

func TestLoadPublicKey_PKCS1(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	der := x509.MarshalPKCS1PublicKey(&priv.PublicKey)
	path := writePEM(t, t.TempDir(), "pub.pem", "RSA PUBLIC KEY", der)

	enc, err := LoadPublicKey(path)
	require.NoError(t, err)
	assert.NotNil(t, enc)
}

func TestLoadPrivateKey_PKCS1(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	der := x509.MarshalPKCS1PrivateKey(priv)
	path := writePEM(t, t.TempDir(), "priv.pem", "RSA PRIVATE KEY", der)

	dec, err := LoadPrivateKey(path)
	require.NoError(t, err)
	assert.NotNil(t, dec)
}

func TestLoadPrivateKey_PKCS8(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	der, err := x509.MarshalPKCS8PrivateKey(priv)
	require.NoError(t, err)
	path := writePEM(t, t.TempDir(), "priv.pem", "PRIVATE KEY", der)

	dec, err := LoadPrivateKey(path)
	require.NoError(t, err)
	assert.NotNil(t, dec)
}

func TestLoadPublicKey_FileMissing(t *testing.T) {
	_, err := LoadPublicKey(filepath.Join(t.TempDir(), "missing.pem"))
	assert.Error(t, err)
}

func TestLoadPublicKey_BadPEM(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.pem")
	require.NoError(t, os.WriteFile(path, []byte("not a pem file"), 0600))
	_, err := LoadPublicKey(path)
	assert.Error(t, err)
}

func TestLoadPrivateKey_FileMissing(t *testing.T) {
	_, err := LoadPrivateKey(filepath.Join(t.TempDir(), "missing.pem"))
	assert.Error(t, err)
}

func TestLoadPrivateKey_BadPEM(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.pem")
	require.NoError(t, os.WriteFile(path, []byte("not a pem file"), 0600))
	_, err := LoadPrivateKey(path)
	assert.Error(t, err)
}

func TestLoadPublicKey_NonRSA(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	require.NoError(t, err)
	path := writePEM(t, t.TempDir(), "pub.pem", "PUBLIC KEY", der)
	_, err = LoadPublicKey(path)
	assert.Error(t, err)
}

func TestLoadPrivateKey_NonRSA(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	require.NoError(t, err)
	path := writePEM(t, t.TempDir(), "priv.pem", "PRIVATE KEY", der)
	_, err = LoadPrivateKey(path)
	assert.Error(t, err)
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	enc := &Encrypter{pub: &priv.PublicKey}
	dec := &Decrypter{priv: priv}

	cases := []struct {
		name string
		size int
	}{
		{"small", 50},
		{"medium", 2000},
		{"large", 64 * 1024},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plaintext := make([]byte, tc.size)
			_, err := rand.Read(plaintext)
			require.NoError(t, err)

			body, encKey, err := enc.Encrypt(plaintext)
			require.NoError(t, err)
			assert.NotEqual(t, plaintext, body)

			decoded, err := dec.Decrypt(body, encKey)
			require.NoError(t, err)
			assert.Equal(t, plaintext, decoded)
		})
	}
}

func TestDecrypt_TamperedBody(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	enc := &Encrypter{pub: &priv.PublicKey}
	dec := &Decrypter{priv: priv}

	body, encKey, err := enc.Encrypt([]byte("hello"))
	require.NoError(t, err)

	body[len(body)-1] ^= 0xFF

	_, err = dec.Decrypt(body, encKey)
	assert.Error(t, err)
}

func TestDecrypt_WrongKey(t *testing.T) {
	priv1, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	priv2, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	enc := &Encrypter{pub: &priv1.PublicKey}
	dec := &Decrypter{priv: priv2}

	body, encKey, err := enc.Encrypt([]byte("hello"))
	require.NoError(t, err)

	_, err = dec.Decrypt(body, encKey)
	assert.Error(t, err)
}

func TestDecrypt_ShortBody(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	enc := &Encrypter{pub: &priv.PublicKey}
	dec := &Decrypter{priv: priv}

	_, encKey, err := enc.Encrypt([]byte("hello"))
	require.NoError(t, err)

	_, err = dec.Decrypt([]byte{1, 2, 3}, encKey)
	assert.Error(t, err)
}
