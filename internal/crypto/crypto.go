package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

type Encrypter struct {
	pub *rsa.PublicKey
}

type Decrypter struct {
	priv *rsa.PrivateKey
}

func LoadPublicKey(path string) (*Encrypter, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("чтение публичного ключа: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("ошибка декодирования PEM блока из публичного ключа")
	}
	if pub, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return &Encrypter{pub: pub}, nil
	}
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("парсинг публичного ключа: %w", err)
	}
	pub, ok := parsed.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("публичный ключ не является RSA ключом: %T", parsed)
	}
	return &Encrypter{pub: pub}, nil
}

func LoadPrivateKey(path string) (*Decrypter, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("чтение приватного ключа: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("ошибка декодирования PEM блока из приватного ключа")
	}
	if priv, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return &Decrypter{priv: priv}, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("парсинг приватного ключа: %w", err)
	}
	priv, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("приватный ключ не является RSA ключом:  %T", parsed)
	}
	return &Decrypter{priv: priv}, nil
}

const aesKeySize = 32

func (e *Encrypter) Encrypt(plaintext []byte) ([]byte, []byte, error) {
	aesKey := make([]byte, aesKeySize)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, nil, fmt.Errorf("генерация aes ключа: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, nil, fmt.Errorf("новый aes шифр: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("новый gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("сгенерировать nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	body := make([]byte, 0, len(nonce)+len(ciphertext))
	body = append(body, nonce...)
	body = append(body, ciphertext...)

	encKey, err := rsa.EncryptPKCS1v15(rand.Reader, e.pub, aesKey)
	if err != nil {
		return nil, nil, fmt.Errorf("rsa шифровка: %w", err)
	}
	return body, encKey, nil
}

func (d *Decrypter) Decrypt(body []byte, encKey []byte) ([]byte, error) {
	aesKey, err := rsa.DecryptPKCS1v15(rand.Reader, d.priv, encKey)
	if err != nil {
		return nil, fmt.Errorf("rsa дешифровка: %w", err)
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("новый aes шифр: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("новый gcm: %w", err)
	}
	if len(body) < gcm.NonceSize() {
		return nil, errors.New("ciphertext слишком короткий")
	}
	nonce := body[:gcm.NonceSize()]
	ciphertext := body[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("aes-gcm open: %w", err)
	}
	return plaintext, nil
}

func NewEncrypterFromKey(pub *rsa.PublicKey) (*Encrypter, error) {
	if pub == nil {
		return nil, errors.New("публичный ключ nil")
	}
	return &Encrypter{pub: pub}, nil
}

func NewDecrypterFromKey(priv *rsa.PrivateKey) (*Decrypter, error) {
	if priv == nil {
		return nil, errors.New("приватный ключ nil")
	}
	return &Decrypter{priv: priv}, nil
}
