package csrf

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"
)

func TestCreate_ValidToken(t *testing.T) {
	tk := NewSimpleToken("mysecret")
	sessionID := "session123"
	expTime := time.Now().Add(time.Minute * 10).Unix()

	token, err := tk.Create(sessionID, expTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	valid, err := tk.Check(sessionID, token)
	if err != nil || !valid {
		t.Errorf("expected token to be valid, got error: %v", err)
	}
}

func TestCreate_InvalidTokenFormat(t *testing.T) {
	tk := NewSimpleToken("mysecret")

	_, err := tk.Check("session123", "invalid.token.format")
	if err == nil {
		t.Errorf("expected error for invalid token format")
	}
}

func TestCreate_InvalidSignature(t *testing.T) {
	tk := NewSimpleToken("mysecret")

	token, _ := tk.Create("session123", time.Now().Add(time.Minute*10).Unix())
	modifiedToken := token + "invalid"

	valid, err := tk.Check("session123", modifiedToken)
	if valid || err == nil {
		t.Errorf("expected invalid signature error")
	}
}

func TestCreate_ExpiredToken(t *testing.T) {
	tk := NewSimpleToken("mysecret")

	expiredTime := time.Now().Add(-time.Minute * 5).Unix()
	token, _ := tk.Create("session123", expiredTime)

	valid, err := tk.Check("session123", token)
	if valid || err == nil || err.Error() != "token expired" {
		t.Errorf("expected token expiration error")
	}
}

func TestCreate_SessionIDMismatch(t *testing.T) {
	tk := NewSimpleToken("mysecret")

	token, _ := tk.Create("session123", time.Now().Add(time.Minute*10).Unix())

	valid, err := tk.Check("wrong_session", token)
	if valid || err == nil || err.Error() != "session ID mismatch" {
		t.Errorf("expected session ID mismatch error")
	}
}

func TestVerify_ValidSignature(t *testing.T) {
	payload := "samplepayload"
	secret := []byte("mysecret")

	// Корректно вычисляем ожидаемую сигнатуру
	expectedSignature := fmt.Sprintf("%x", sha256.Sum256(append([]byte(payload), secret...)))

	if !Verify(payload, expectedSignature, secret) {
		t.Errorf("expected signature to be valid")
	}
}

func TestVerify_InvalidSignature(t *testing.T) {
	payload := "samplepayload"
	secret := []byte("wrongsecret")
	signature := "c94e56f5c86f98392480d93fbee3ad6f5e5c5aeeb4f2f232ec30b34a6bd64f95"

	if Verify(payload, signature, secret) {
		t.Errorf("expected signature to be invalid")
	}
}
