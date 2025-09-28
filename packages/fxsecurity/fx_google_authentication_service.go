package fxsecurity

import (
	"bytes"
	"encoding/base32"
	"image/png"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

//Make sure your local time zone is correct

type IGoogleAuthenticationService interface {
	GenerateKey(accountName string) (*otp.Key, error)
	VerifyOTP(accountName string, otpCode string) bool
	GenerateOTP(accountName string) ([]byte, error)
}

type googleAuthenticationService struct {
	issuer string
	secret string
}

func NewGoogleAuthenticationService(issuer string, secret string) IGoogleAuthenticationService {
	return &googleAuthenticationService{issuer: issuer, secret: secret}
}

func (this *googleAuthenticationService) GenerateKey(accountName string) (*otp.Key, error) {
	secretKey := this.issuer + "_" + this.secret + "_" + accountName
	encoder := base32.StdEncoding
	secretBase32Encoding := encoder.EncodeToString([]byte(secretKey))
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      this.issuer,
		AccountName: accountName,
		Secret:      []byte(secretBase32Encoding), // (Base32-encoded)
	})
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (this *googleAuthenticationService) VerifyOTP(accountName string, otpCode string) bool {
	key, err := this.GenerateKey(accountName)
	if err != nil {
		return false
	}
	secret := key.Secret()
	isValid := totp.Validate(otpCode, secret)
	return isValid
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
	png.Encode(&buf, img)
	return buf.Bytes(), nil
}
