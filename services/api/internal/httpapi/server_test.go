package httpapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/0xmarkhydra/FreeEducation/services/api/internal/auth"
	"github.com/0xmarkhydra/FreeEducation/services/api/internal/catalog"
)

const secret = "00000000000000000000000000000000"
const testID = "11111111-1111-4111-8111-111111111111"

type fakeStore struct {
	filter       catalog.Filter
	actor        catalog.Actor
	created      bool
	transitioned bool
}

func (f *fakeStore) List(_ context.Context, filter catalog.Filter) ([]catalog.Book, error) {
	f.filter = filter
	return []catalog.Book{}, nil
}
func (f *fakeStore) Get(_ context.Context, _, _ string) (catalog.Book, error) {
	return catalog.Book{}, catalog.ErrNotFound
}
func (f *fakeStore) Create(_ context.Context, a catalog.Actor, b catalog.BookInput) (catalog.Book, error) {
	f.actor = a
	if err := b.Validate(a); err != nil {
		return catalog.Book{}, err
	}
	f.created = true
	return catalog.Book{ID: testID, BookInput: b, Status: "draft"}, nil
}
func (f *fakeStore) Transition(_ context.Context, a catalog.Actor, _, from, to string) error {
	f.actor = a
	f.transitioned = true
	if from != "review" || to != "approved" {
		return catalog.ErrConflict
	}
	return nil
}
func makeJWT(t *testing.T, school string, roles []string) string {
	t.Helper()
	h, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	p, _ := json.Marshal(auth.Claims{Subject: testID, SchoolID: school, Issuer: "issuer", Audience: "api", Expires: time.Now().Add(time.Hour).Unix(), Roles: roles})
	data := base64.RawURLEncoding.EncodeToString(h) + "." + base64.RawURLEncoding.EncodeToString(p)
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(data))
	return data + "." + base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}
func serveTest(t *testing.T, f *fakeStore) http.Handler {
	t.Helper()
	v, err := auth.NewVerifier(secret, "issuer", "api")
	if err != nil {
		t.Fatal(err)
	}
	return New(f, v, nil)
}
func TestAnonymousAccessAndSpoofedSchool(t *testing.T) {
	f := &fakeStore{}
	s := serveTest(t, f)
	r := httptest.NewRequest("GET", "/v1/books", nil)
	r.Header.Set("X-School-ID", "attacker-selected-school")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 200 || f.filter.SchoolID != "" {
		t.Fatalf("unauthenticated tenant leakage: %d %#v", w.Code, f.filter)
	}
	post := httptest.NewRequest("POST", "/v1/admin/books", strings.NewReader("{}"))
	w = httptest.NewRecorder()
	s.ServeHTTP(w, post)
	if w.Code != 401 || f.created {
		t.Fatalf("unauthenticated write: %d", w.Code)
	}
}
func TestClaimsScopedToSchoolAndRoles(t *testing.T) {
	f := &fakeStore{}
	s := serveTest(t, f)
	r := httptest.NewRequest("GET", "/v1/books?grade=6&limit=2", nil)
	r.Header.Set("Authorization", "Bearer "+makeJWT(t, testID, []string{"student"}))
	r.Header.Set("X-School-ID", "spoofed")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 200 || f.filter.SchoolID != testID || f.filter.Grade != 6 || f.filter.Limit != 2 {
		t.Fatalf("invalid scoped filter: %#v / %d", f.filter, w.Code)
	}
	input := "{\"title\":\"Toan 6\",\"grade\":6,\"subject_code\":\"math\",\"rights_owner\":\"Owner\",\"rights_license\":\"Permit\",\"rights_source\":\"Publisher\",\"visibility\":\"school\"}"
	r = httptest.NewRequest("POST", "/v1/admin/books", strings.NewReader(input))
	r.Header.Set("Authorization", "Bearer "+makeJWT(t, testID, []string{"student"}))
	w = httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatalf("student allowed creation: %d", w.Code)
	}
	r = httptest.NewRequest("POST", "/v1/admin/books", strings.NewReader(input))
	r.Header.Set("Authorization", "Bearer "+makeJWT(t, testID, []string{"school_admin"}))
	w = httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 201 || !f.created || f.actor.SchoolID != testID {
		t.Fatalf("school admin creation rejected: %d %#v", w.Code, f.actor)
	}
}
func TestReviewerWorkflow(t *testing.T) {
	f := &fakeStore{}
	s := serveTest(t, f)
	r := httptest.NewRequest("POST", "/v1/admin/books/"+testID+"/approve", nil)
	r.Header.Set("Authorization", "Bearer "+makeJWT(t, "", []string{"content_reviewer"}))
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 200 || !f.transitioned {
		t.Fatalf("reviewer denied: %d", w.Code)
	}
	r = httptest.NewRequest("POST", "/v1/admin/books/"+testID+"/approve", nil)
	r.Header.Set("Authorization", "Bearer "+makeJWT(t, testID, []string{"student"}))
	w = httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatalf("student approval accepted: %d", w.Code)
	}
}
func TestMalformedTokenAndGrade(t *testing.T) {
	f := &fakeStore{}
	s := serveTest(t, f)
	r := httptest.NewRequest("GET", "/v1/books", nil)
	r.Header.Set("Authorization", "Bearer invalid")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("malformed token accepted: %d", w.Code)
	}
	r = httptest.NewRequest("GET", "/v1/books?grade=13", nil)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatalf("invalid grade accepted: %d", w.Code)
	}
}
