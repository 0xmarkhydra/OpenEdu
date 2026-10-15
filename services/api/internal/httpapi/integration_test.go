package httpapi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/0xmarkhydra/OpenEdu/services/api/internal/auth"
	"github.com/0xmarkhydra/OpenEdu/services/api/internal/catalog"
	"github.com/0xmarkhydra/OpenEdu/services/api/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Run exclusively against a disposable development database named openedu_test.
// Never point OPENEDU_TEST_DSN at Railway production or an existing user database.
func TestIdentityAndSchoolIsolationPostgres(t *testing.T) {
	dsn := os.Getenv("OPENEDU_TEST_DSN")
	if dsn == "" {
		t.Skip("OPENEDU_TEST_DSN not configured")
	}
	if !strings.Contains(dsn, "openedu_test") {
		t.Fatal("refusing non-test database")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := migrations.Apply(ctx, db); err != nil {
		t.Fatal(err)
	}
	suffix := time.Now().Format("150405.000000")
	var schoolA, schoolB, userID string
	for i, school := range []*string{&schoolA, &schoolB} {
		if err := db.QueryRowContext(ctx, "INSERT INTO schools(name,slug) VALUES($1,$2) RETURNING id::text", "Test School", strings.ReplaceAll("test-school-"+suffix+"-"+string(rune('a'+i)), ".", "")).Scan(school); err != nil {
			t.Fatal(err)
		}
	}
	email := "integration-" + strings.ReplaceAll(suffix, ".", "") + "@example.test"
	if err := db.QueryRowContext(ctx, "INSERT INTO users(email,display_name) VALUES($1,'Integration learner') RETURNING id::text", email).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	password := "integration-strong-2026"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, "INSERT INTO user_credentials(user_id,password_hash) VALUES($1::uuid,$2)", userID, hash); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ExecContext(ctx, "INSERT INTO school_memberships(school_id,user_id,role) VALUES($1::uuid,$2::uuid,'student')", schoolA, userID); err != nil {
		t.Fatal(err)
	}
	v, err := auth.NewVerifier("0123456789abcdef0123456789abcdef", "issuer", "audience")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(New(catalog.NewPostgresStore(db), v, db))
	defer server.Close()
	post := func(path string, data any) *http.Response {
		b, _ := json.Marshal(data)
		res, err := http.Post(server.URL+path, "application/json", bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	login := post("/v1/auth/login", map[string]string{"email": email, "password": password})
	if login.StatusCode != 200 {
		t.Fatalf("login status %d", login.StatusCode)
	}
	var session auth.LoginResult
	if err := json.NewDecoder(login.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	login.Body.Close()
	if session.SchoolID != schoolA || len(session.Roles) != 1 || session.Roles[0] != "student" {
		t.Fatalf("unexpected session: %+v", session)
	}
	createBook := func(school string) string {
		var id string
		err := db.QueryRowContext(ctx, `INSERT INTO books(school_id,title,grade,subject_code,rights_owner,rights_license,rights_source,visibility,status,created_by)
VALUES($1::uuid,'Integration Test Book',5,'test','test owner','test-only','test fixture','school','published',$2::uuid) RETURNING id::text`, school, userID).Scan(&id)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	a, b := createBook(schoolA), createBook(schoolB)
	req, _ := http.NewRequest("GET", server.URL+"/v1/books", nil)
	req.Header.Set("Authorization", "Bearer "+session.AccessToken)
	req.Header.Set("X-School-ID", schoolB)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var books struct {
		Items []catalog.Book `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&books); err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	hasA, hasB := false, false
	for _, book := range books.Items {
		if book.ID == a {
			hasA = true
		}
		if book.ID == b {
			hasB = true
		}
	}
	if !hasA || hasB {
		t.Fatalf("cross-school book leakage: a=%v b=%v", hasA, hasB)
	}
	req, _ = http.NewRequest("GET", server.URL+"/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+session.AccessToken)
	if _, err = db.ExecContext(ctx, "UPDATE school_memberships SET is_active=false WHERE user_id=$1::uuid AND school_id=$2::uuid", userID, schoolA); err != nil {
		t.Fatal(err)
	}
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 401 {
		t.Fatalf("revoked membership accepted: %d", res.StatusCode)
	}

	// A revoked school role does not block a separately granted platform role.
	if _, err = db.ExecContext(ctx, `INSERT INTO platform_roles(user_id,role) VALUES($1::uuid,'platform_admin')`, userID); err != nil {
		t.Fatal(err)
	}
	login = post("/v1/auth/login", map[string]string{"email": email, "password": password})
	if login.StatusCode != 200 {
		t.Fatalf("platform login: %d", login.StatusCode)
	}
	session = auth.LoginResult{}
	if err := json.NewDecoder(login.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	login.Body.Close()
	if session.SchoolID != "" || len(session.Roles) != 1 || session.Roles[0] != "platform_admin" {
		t.Fatalf("platform role: school=%q roles=%v", session.SchoolID, session.Roles)
	}
	authPost := func(path string, body any, token string) *http.Response {
		data, _ := json.Marshal(body)
		req, err := http.NewRequest("POST", server.URL+path, bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		r, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	slug := "new-integration-" + strings.ReplaceAll(suffix, ".", "")
	newSchool := authPost("/v1/admin/schools", map[string]string{"name": "Integration campus", "slug": slug}, session.AccessToken)
	if newSchool.StatusCode != 201 {
		t.Fatalf("new school denied: %d", newSchool.StatusCode)
	}
	var created School
	if err := json.NewDecoder(newSchool.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	newSchool.Body.Close()
	newMemberEmail := "member-" + strings.ReplaceAll(suffix, ".", "") + "@example.test"
	newMember := authPost("/v1/admin/schools/"+created.ID+"/members", map[string]string{"email": newMemberEmail, "name": "Integration administrator", "password": "member-password-2026", "role": "school_admin"}, session.AccessToken)
	if newMember.StatusCode != 201 {
		t.Fatalf("member creation denied: %d", newMember.StatusCode)
	}
	newMember.Body.Close()
	memberSession := post("/v1/auth/login", map[string]string{"email": newMemberEmail, "password": "member-password-2026", "school_id": created.ID})
	if memberSession.StatusCode != 200 {
		t.Fatalf("school admin login: %d", memberSession.StatusCode)
	}
	var member auth.LoginResult
	if err := json.NewDecoder(memberSession.Body).Decode(&member); err != nil {
		t.Fatal(err)
	}
	memberSession.Body.Close()
	if member.SchoolID != created.ID || member.Roles[0] != "school_admin" {
		t.Fatalf("new member identity wrong: %+v", member)
	}
	denied := authPost("/v1/admin/schools", map[string]string{"name": "Forbidden", "slug": "forbidden-school"}, member.AccessToken)
	denied.Body.Close()
	if denied.StatusCode != 403 {
		t.Fatalf("school role escaped to platform admin: %d", denied.StatusCode)
	}

}
