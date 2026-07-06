package fxsecurity

import (
	"bytes"
	"fmt"
	"image/png"
	"net/url"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

//Make sure your local time zone is correct

type IGoogleAuthenticationService interface {
	GenerateKey(accountName string) (*otp.Key, error)
	VerifyOTP(accountName string, otpCode string) bool
	GenerateOTP(accountName string) ([]byte, error)
	// IsEnrolled reports whether the account has a TOTP secret set up.
	IsEnrolled(accountName string) bool
	// DisableOTP removes the account's secret (reset / un-enroll).
	DisableOTP(accountName string) error
}

type googleAuthenticationService struct {
	issuer string
	// store holds a random per-account secret. Each account is assigned its own
	// secret on enrollment so a compromise or reset affects only that account.
	store ISecretStore
}

// NewGoogleAuthenticationService builds the service over a secret store: each
// account is assigned a random secret on first enrollment and persisted, so
// secrets can be reset per account and enrollment state is tracked.
func NewGoogleAuthenticationService(issuer string, store ISecretStore) IGoogleAuthenticationService {
	if store == nil {
		panic("fxsecurity: NewGoogleAuthenticationService requires a non-nil ISecretStore")
	}
	return &googleAuthenticationService{issuer: issuer, store: store}
}

// GenerateKey returns the account's TOTP key: the existing key if the account is
// already enrolled, otherwise a freshly generated random secret that is persisted
// and returned (enrollment).
func (this *googleAuthenticationService) GenerateKey(accountName string) (*otp.Key, error) {
	secret, found, err := this.store.GetSecret(accountName)
	if err != nil {
		return nil, err
	}
	if found {
		return this.keyFromSecret(accountName, secret)
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      this.issuer,
		AccountName: accountName,
	})
	if err != nil {
		return nil, err
	}
	if err := this.store.SaveSecret(accountName, key.Secret()); err != nil {
		return nil, err
	}
	return key, nil
}

func (this *googleAuthenticationService) VerifyOTP(accountName string, otpCode string) bool {
	secret, found, err := this.store.GetSecret(accountName)
	if err != nil || !found {
		return false
	}
	return totp.Validate(otpCode, secret)
}

func (this *googleAuthenticationService) GenerateOTP(accountName string) ([]byte, error) {
	key, err := this.GenerateKey(accountName)
	if err != nil {
		return nil, err
	}
	// Convert TOTP key into a PNG
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return nil, err
	}
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (this *googleAuthenticationService) IsEnrolled(accountName string) bool {
	exists, err := this.store.Exists(accountName)
	if err != nil {
		return false
	}
	return exists
}

func (this *googleAuthenticationService) DisableOTP(accountName string) error {
	return this.store.DeleteSecret(accountName)
}

// ----------------------------------------------------------------------------------------
// Private helpers
// ----------------------------------------------------------------------------------------

// keyFromSecret rebuilds a TOTP key object from a stored base32 secret so the QR
// image and provisioning URI can be regenerated for an already-enrolled account.
func (this *googleAuthenticationService) keyFromSecret(accountName string, secret string) (*otp.Key, error) {
	otpauth := fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s",
		url.PathEscape(this.issuer),
		url.PathEscape(accountName),
		secret,
		url.QueryEscape(this.issuer),
	)
	return otp.NewKeyFromURL(otpauth)
}
