package usecase

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type CryptToken struct {
	Secret []byte
}

type TokenData struct {
	SessionID string `json:"session_id"`
	Exp       int64  `json:"exp"`
}

func NewSimpleToken(secret string) *CryptToken {
	return &CryptToken{Secret: []byte(secret)}
}

func (tk *CryptToken) Create(sessionID string, tokenExpTime int64) (string, error) {
	td := &TokenData{SessionID: sessionID, Exp: tokenExpTime}
	data, err := json.Marshal(td)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token %w", err)
	}
	payload := base64.StdEncoding.EncodeToString(data)

	signature := fmt.Sprintf("%x", sha256.Sum256(append([]byte(payload), tk.Secret...)))
	token := payload + "." + signature

	return token, nil
}

func (tk *CryptToken) Check(sid string, inputToken string) (bool, error) {

	parts := strings.Split(inputToken, ".")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid token format")
	}
	payload, signature := parts[0], parts[1]

	if !Verify(payload, signature, tk.Secret) {
		return false, fmt.Errorf("invalid token signature")
	}

	data, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return false, fmt.Errorf("failed to decode base64 payload: %w", err)
	}

	td := TokenData{}
	err = json.Unmarshal(data, &td)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal token %w", err)
	}

	if td.Exp < time.Now().Unix() {
		return false, fmt.Errorf("token expired")
	}

	if td.SessionID != sid {
		return false, fmt.Errorf("session ID mismatch")
	}

	return true, nil
}

func Verify(payload, signature string, secret []byte) bool {

	expectedSignature := fmt.Sprintf("%x", sha256.Sum256(append([]byte(payload), secret...)))
	return signature == expectedSignature
}
