package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func GenerateAPIKey() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "rsk_live_" + base64.RawURLEncoding.EncodeToString(buf), nil
}

func HashAPIKey(secret, key string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(key))
	return hex.EncodeToString(mac.Sum(nil))
}

func ValidateKeyPrefix(key string) error {
	if len(key) < len("rsk_live_")+16 || key[:len("rsk_live_")] != "rsk_live_" {
		return fmt.Errorf("invalid api key format")
	}
	return nil
}
