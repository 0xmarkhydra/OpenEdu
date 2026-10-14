package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

const testSecret = "1234567890abcdef1234567890abcdef"

func sign(t *testing.T, c Claims, alg string) string {
	t.Helper()
	h, _ := json.Marshal(map[string]string{"alg": alg, "typ": "JWT"})
	b, _ := json.Marshal(c)
	payload := base64.RawURLEncoding.EncodeToString(h) + "." + base64.RawURLEncoding.EncodeToString(b)
	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write([]byte(payload))
	return payload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
func TestVerify(t *testing.T) {
	v, err := NewVerifier(testSecret, "freeeducation-identity", "freeeducation-api")
	if err != nil {
		t.Fatal(err)
	}
	base := Claims{Subject: "11111111-1111-4111-8111-111111111111", Issuer: "freeeducation-identity", Audience: "freeeducation-api", Expires: time.Now().Add(time.Hour).Unix(), Roles: []string{"student"}}
	token := sign(t, base, "HS256")
	claims, err := v.Verify(token)
	if err != nil || !claims.HasRole("student") {
		t.Fatal(err)
	}
	for _, s := range []string{token + "x", sign(t, base, "none")} {
		if _, err := v.Verify(s); !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("expected rejection: %v", err)
		}
	}
	base.Expires = time.Now().Add(-time.Hour).Unix()
	if _, err := v.Verify(sign(t, base, "HS256")); !errors.Is(err, ErrInvalidToken) {
		t.Fatal("expired token accepted")
	}
	if _, err := NewVerifier("short", "issuer", "audience"); err == nil {
		t.Fatal("short secret accepted")
	}
}
