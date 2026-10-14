package auth

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
)

var ErrPassword = errors.New("password must have at least 12 and at most 72 bytes")

// HashPassword uses bcrypt cost 12 with a random per-password salt.
func HashPassword(password string) (string, error) {
	if len(password) < 12 || len(password) > 72 {
		return "", ErrPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func VerifyPassword(encoded, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(encoded), []byte(password)) == nil
}
