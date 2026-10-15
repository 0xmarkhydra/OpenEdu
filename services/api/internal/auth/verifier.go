package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"
)

var ErrInvalidToken = errors.New("invalid authentication token")

var uuidID = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

type Claims struct {
	Subject     string   `json:"sub"`
	SchoolID    string   `json:"school_id"`
	Roles       []string `json:"roles"`
	Issuer      string   `json:"iss"`
	Audience    string   `json:"aud"`
	Expires     int64    `json:"exp"`
	NotBefore   int64    `json:"nbf"`
	AuthVersion int64    `json:"auth_version"`
}

func (c Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}
func (c Claims) HasAny(roles ...string) bool {
	for _, r := range roles {
		if c.HasRole(r) {
			return true
		}
	}
	return false
}

type Verifier struct {
	secret   []byte
	issuer   string
	audience string
}

func NewVerifier(secret, issuer, audience string) (*Verifier, error) {
	if len(secret) < 32 || issuer == "" || audience == "" {
		return nil, errors.New("JWT_SECRET must be >=32 bytes with explicit issuer and audience")
	}
	return &Verifier{secret: []byte(secret), issuer: issuer, audience: audience}, nil
}

func (v *Verifier) Verify(token string) (Claims, error) {
	var c Claims
	if len(token) > 8192 {
		return c, ErrInvalidToken
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return c, ErrInvalidToken
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return c, ErrInvalidToken
	}
	var header struct {
		Algorithm string `json:"alg"`
		Type      string `json:"typ"`
	}
	if json.Unmarshal(headerBytes, &header) != nil || header.Algorithm != "HS256" {
		return c, ErrInvalidToken
	}
	mac := hmac.New(sha256.New, v.secret)
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(mac.Sum(nil), sig) {
		return c, ErrInvalidToken
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || json.Unmarshal(data, &c) != nil {
		return Claims{}, ErrInvalidToken
	}
	now := time.Now().Unix()
	if !uuidID.MatchString(c.Subject) || (c.SchoolID != "" && !uuidID.MatchString(c.SchoolID)) || c.Issuer != v.issuer || c.Audience != v.audience || c.Expires <= now || c.NotBefore > now || c.Expires > now+86400 {
		return Claims{}, ErrInvalidToken
	}
	if len(c.Roles) == 0 {
		return Claims{}, ErrInvalidToken
	}
	return c, nil
}

func (v *Verifier) Issue(c Claims, ttl time.Duration) (string, error) {
	if !uuidID.MatchString(c.Subject) || (c.SchoolID != "" && !uuidID.MatchString(c.SchoolID)) || len(c.Roles) == 0 || ttl <= 0 || ttl > time.Hour {
		return "", ErrInvalidToken
	}
	if c.AuthVersion == 0 {
		c.AuthVersion = 1
	}
	c.Issuer = v.issuer
	c.Audience = v.audience
	c.NotBefore = time.Now().Unix()
	c.Expires = time.Now().Add(ttl).Unix()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	body, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	unsigned := header + "." + base64.RawURLEncoding.EncodeToString(body)
	mac := hmac.New(sha256.New, v.secret)
	mac.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}
