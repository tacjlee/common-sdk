package fxsecurity

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

// Encrypted store round-trips: what lands in the inner store is ciphertext, but
// GetSecret returns the original plaintext.
func TestEncryptedSecretStoreRoundTrip(t *testing.T) {
	inner := NewInMemorySecretStore()
	enc, err := NewEncryptedSecretStore(inner, "test-encryption-key")
	if err != nil {
		t.Fatalf("build encrypted store: %v", err)
	}
	if err := enc.SaveSecret("alice", "JBSWY3DPEHPK3PXP"); err != nil {
		t.Fatalf("save: %v", err)
	}

	stored, _, _ := inner.GetSecret("alice")
	if stored == "JBSWY3DPEHPK3PXP" {
		t.Fatal("secret stored in plaintext; expected ciphertext")
	}
	got, found, err := enc.GetSecret("alice")
	if err != nil || !found || got != "JBSWY3DPEHPK3PXP" {
		t.Fatalf("get: got=%q found=%v err=%v", got, found, err)
	}
}

// The constructor rejects a nil store rather than deferring the failure to first use.
func TestNilStorePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil store")
		}
	}()
	NewGoogleAuthenticationService("dxo", nil)
}

// Store-backed service: enroll assigns and persists a random secret; a code
// generated from that secret verifies; DisableOTP un-enrolls.
func TestStoreBackedEnrollVerifyDisable(t *testing.T) {
	svc := NewGoogleAuthenticationService("dxo", NewInMemorySecretStore())

	if svc.IsEnrolled("bob") {
		t.Fatal("bob should not be enrolled yet")
	}

	key, err := svc.GenerateKey("bob") // enrollment
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	if !svc.IsEnrolled("bob") {
		t.Fatal("bob should be enrolled after GenerateKey")
	}

	code, err := totp.GenerateCode(key.Secret(), time.Now())
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}
	if !svc.VerifyOTP("bob", code) {
		t.Fatal("valid code did not verify")
	}
	if svc.VerifyOTP("bob", "000000") {
		t.Fatal("bogus code verified")
	}

	if err := svc.DisableOTP("bob"); err != nil {
		t.Fatalf("disable: %v", err)
	}
	if svc.IsEnrolled("bob") {
		t.Fatal("bob should be un-enrolled after DisableOTP")
	}
	if svc.VerifyOTP("bob", code) {
		t.Fatal("verify should fail after disable")
	}
}
