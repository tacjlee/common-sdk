package fxsecurity

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

// encryptedSecretStore wraps another ISecretStore and encrypts the secret at rest
// with AES-256-GCM, so a database dump never exposes usable TOTP secrets. Only the
// secret value is encrypted; the account key stays clear so lookups still work.
type encryptedSecretStore struct {
	inner ISecretStore
	gcm   cipher.AEAD
}

// NewEncryptedSecretStore wraps inner so secrets are encrypted before storage and
// decrypted on read. The AES-256 key is derived from encryptionKey via SHA-256.
func NewEncryptedSecretStore(inner ISecretStore, encryptionKey string) (ISecretStore, error) {
	keyDigest := sha256.Sum256([]byte(encryptionKey))
	block, err := aes.NewCipher(keyDigest[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &encryptedSecretStore{inner: inner, gcm: gcm}, nil
}

func (this *encryptedSecretStore) SaveSecret(accountName string, secret string) error {
	nonce := make([]byte, this.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	sealed := this.gcm.Seal(nonce, nonce, []byte(secret), nil)
	return this.inner.SaveSecret(accountName, base64.StdEncoding.EncodeToString(sealed))
}

func (this *encryptedSecretStore) GetSecret(accountName string) (string, bool, error) {
	stored, found, err := this.inner.GetSecret(accountName)
	if err != nil || !found {
		return "", found, err
	}
	sealed, err := base64.StdEncoding.DecodeString(stored)
	if err != nil {
		return "", false, err
	}
	nonceSize := this.gcm.NonceSize()
	if len(sealed) < nonceSize {
		return "", false, errors.New("fxsecurity: ciphertext too short")
	}
	nonce, ciphertext := sealed[:nonceSize], sealed[nonceSize:]
	plain, err := this.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", false, err
	}
	return string(plain), true, nil
}

func (this *encryptedSecretStore) DeleteSecret(accountName string) error {
	return this.inner.DeleteSecret(accountName)
}

func (this *encryptedSecretStore) Exists(accountName string) (bool, error) {
	return this.inner.Exists(accountName)
}
