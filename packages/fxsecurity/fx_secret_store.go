package fxsecurity

import (
	"sync"

	"github.com/tacjlee/common-sdk/packages/fxrepository"
)

// ISecretStore persists a per-account TOTP secret. It replaces the deterministic
// secret derivation used when no store is configured: instead of recomputing a
// secret from a single master key, each account gets its own random secret that
// can be individually reset (lost device) without affecting anyone else, and whose
// presence records that the account has enrolled.
type ISecretStore interface {
	// SaveSecret stores (or replaces) the base32 TOTP secret for an account.
	SaveSecret(accountName string, secret string) error
	// GetSecret returns the stored secret for an account. found is false when the
	// account has no secret (not enrolled).
	GetSecret(accountName string) (secret string, found bool, err error)
	// DeleteSecret removes an account's secret (un-enroll / reset). Idempotent.
	DeleteSecret(accountName string) error
	// Exists reports whether an account has a stored secret (is enrolled).
	Exists(accountName string) (bool, error)
}

// ----------------------------------------------------------------------------------------
// In-memory store (tests / single-instance dev). Not durable across restarts.
// ----------------------------------------------------------------------------------------

type inMemorySecretStore struct {
	mu      sync.RWMutex
	secrets map[string]string
}

func NewInMemorySecretStore() ISecretStore {
	return &inMemorySecretStore{secrets: make(map[string]string)}
}

func (this *inMemorySecretStore) SaveSecret(accountName string, secret string) error {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.secrets[accountName] = secret
	return nil
}

func (this *inMemorySecretStore) GetSecret(accountName string) (string, bool, error) {
	this.mu.RLock()
	defer this.mu.RUnlock()
	secret, found := this.secrets[accountName]
	return secret, found, nil
}

func (this *inMemorySecretStore) DeleteSecret(accountName string) error {
	this.mu.Lock()
	defer this.mu.Unlock()
	delete(this.secrets, accountName)
	return nil
}

func (this *inMemorySecretStore) Exists(accountName string) (bool, error) {
	this.mu.RLock()
	defer this.mu.RUnlock()
	_, found := this.secrets[accountName]
	return found, nil
}

// ----------------------------------------------------------------------------------------
// GORM-backed store (durable). Uses the SDK's generic repository.
// ----------------------------------------------------------------------------------------

// GoogleAuthenticatorSecret is the persistence model for a per-account TOTP secret.
//
// NOTE: the secret is equivalent to a shared password. Protect it at rest —
// enable database/column encryption, or wrap this store in an encrypting decorator
// (encrypt before SaveSecret, decrypt after GetSecret).
type GoogleAuthenticatorSecret struct {
	AccountName string `gorm:"column:account_name;primaryKey;size:255"`
	Secret      string `gorm:"column:secret;size:512"`
}

func (GoogleAuthenticatorSecret) TableName() string {
	return "google_authenticator_secret"
}

type gormSecretStore struct {
	repository fxrepository.IGenericRepository
}

// NewGormSecretStore builds a durable secret store over the SDK generic repository.
// Ensure the google_authenticator_secret table exists (e.g. via AutoMigrate on the
// GoogleAuthenticatorSecret model).
func NewGormSecretStore(repository fxrepository.IGenericRepository) ISecretStore {
	return &gormSecretStore{repository: repository}
}

func (this *gormSecretStore) SaveSecret(accountName string, secret string) error {
	record := &GoogleAuthenticatorSecret{AccountName: accountName, Secret: secret}
	_, err := this.repository.Save(record) // upsert on primary key
	return err
}

func (this *gormSecretStore) GetSecret(accountName string) (string, bool, error) {
	var record GoogleAuthenticatorSecret
	result := this.repository.GetDB().
		Where("account_name = ?", accountName).
		Limit(1).
		Find(&record)
	if result.Error != nil {
		return "", false, result.Error
	}
	if result.RowsAffected == 0 {
		return "", false, nil
	}
	return record.Secret, true, nil
}

func (this *gormSecretStore) DeleteSecret(accountName string) error {
	_, err := this.repository.Delete(&GoogleAuthenticatorSecret{}, "account_name = ?", accountName)
	return err
}

func (this *gormSecretStore) Exists(accountName string) (bool, error) {
	_, found, err := this.GetSecret(accountName)
	return found, err
}
