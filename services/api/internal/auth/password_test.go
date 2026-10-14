package auth

import "testing"

func TestPasswordHashing(t *testing.T) {
	a, err := HashPassword("secure-passphrase-2026")
	if err != nil {
		t.Fatal(err)
	}
	b, err := HashPassword("secure-passphrase-2026")
	if err != nil || a == b {
		t.Fatal("salt not unique")
	}
	if !VerifyPassword(a, "secure-passphrase-2026") || VerifyPassword(a, "invalid-password") {
		t.Fatal("password check failed")
	}
	if _, err := HashPassword("weak"); err == nil {
		t.Fatal("short password accepted")
	}
	if _, err := HashPassword(string(make([]byte, 73))); err == nil {
		t.Fatal("oversized password accepted")
	}
}
