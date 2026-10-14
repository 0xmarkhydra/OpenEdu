package auth

import (
	"testing"
	"time"
)

func TestIssuerRoundTrip(t *testing.T) {
	v, err := NewVerifier("1234567890abcdef1234567890abcdef", "issuer", "audience")
	if err != nil {
		t.Fatal(err)
	}
	c := Claims{Subject: "11111111-1111-4111-8111-111111111111", SchoolID: "22222222-2222-4222-8222-222222222222", Roles: []string{"teacher"}}
	token, err := v.Issue(c, 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := v.Verify(token)
	if err != nil || decoded.Subject != c.Subject || decoded.SchoolID != c.SchoolID || !decoded.HasRole("teacher") {
		t.Fatalf("invalid session: %#v %v", decoded, err)
	}
	if _, err := v.Issue(c, 25*time.Hour); err == nil {
		t.Fatal("unbounded session")
	}
	if _, err := v.Issue(Claims{Subject: "invalid", Roles: []string{"teacher"}}, time.Minute); err == nil {
		t.Fatal("malformed subject")
	}
}
