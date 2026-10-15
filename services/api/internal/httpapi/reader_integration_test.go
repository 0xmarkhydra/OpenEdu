package httpapi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

func TestEditorialReaderPublishingPostgres(t *testing.T) {
	dsn := os.Getenv("OPENEDU_TEST_DSN")
	if dsn == "" {
		t.Skip("no disposable PostgreSQL DSN")
	}
	if !strings.Contains(dsn, "openedu_test") || strings.Contains(dsn, "railway") {
		t.Fatal("not a disposable DB")
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
	suffix := fmt.Sprintf("reader-%d", time.Now().UnixNano())
	var school string
	if err := db.QueryRowContext(ctx, "INSERT INTO schools(name,slug) VALUES('Editorial Test',$1) RETURNING id::text", suffix).Scan(&school); err != nil {
		t.Fatal(err)
	}
	ids := map[string]string{}
	for _, role := range []string{"school_admin", "content_reviewer", "student"} {
		var id string
		if err := db.QueryRowContext(ctx, "INSERT INTO users(email,display_name) VALUES($1,$2) RETURNING id::text", role+suffix+"@example.test", role).Scan(&id); err != nil {
			t.Fatal(err)
		}
		if _, err := db.ExecContext(ctx, "INSERT INTO school_memberships(school_id,user_id,role) VALUES($1::uuid,$2::uuid,$3)", school, id, role); err != nil {
			t.Fatal(err)
		}
		ids[role] = id
	}
	verifier, err := auth.NewVerifier("0123456789abcdef0123456789abcdef", "test", "api")
	if err != nil {
		t.Fatal(err)
	}
	tokens := map[string]string{}
	for role, id := range ids {
		token, err := verifier.Issue(auth.Claims{Subject: id, SchoolID: school, Roles: []string{role}}, time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		tokens[role] = token
	}
	app := httptest.NewServer(New(catalog.NewPostgresStore(db), verifier, db))
	defer app.Close()
	request := func(role, method, path string, body any) (int, map[string]any) {
		var buf []byte
		if body != nil {
			buf, _ = json.Marshal(body)
		}
		req, err := http.NewRequestWithContext(ctx, method, app.URL+path, bytes.NewReader(buf))
		if err != nil {
			t.Fatal(err)
		}
		if role != "" {
			req.Header.Set("Authorization", "Bearer "+tokens[role])
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var data map[string]any
		if json.NewDecoder(resp.Body).Decode(&data) != nil {
			t.Fatal("non JSON")
		}
		return resp.StatusCode, data
	}
	check := func(code, got int, data map[string]any) {
		if got != code {
			t.Fatalf("want %d got %d %+v", code, got, data)
		}
	}
	status, book := request("school_admin", "POST", "/v1/admin/books", map[string]any{"title": "Synthetic licensed test lesson", "grade": 6, "subject_code": "test", "rights_owner": "Synthetic fixture", "rights_license": "test-only", "rights_source": "generated fixture", "visibility": "school"})
	check(201, status, book)
	bookID := book["id"].(string)
	status, _ = request("student", "GET", "/v1/books/"+bookID+"/outline", nil)
	check(404, status, nil)
	status, chap := request("school_admin", "POST", "/v1/admin/books/"+bookID+"/chapters", map[string]any{"title": "Chapter One", "position": 1})
	check(201, status, chap)
	status, _ = request("school_admin", "GET", "/v1/admin/books/"+bookID+"/chapters", nil)
	check(200, status, nil)
	status, lesson := request("school_admin", "POST", "/v1/admin/chapters/"+chap["id"].(string)+"/lessons", map[string]any{"title": "Safe test chapter", "position": 1, "content": "Synthetic educational content only."})
	check(201, status, lesson)
	status, _ = request("school_admin", "POST", "/v1/admin/books/"+bookID+"/submit", map[string]any{})
	check(200, status, nil)
	status, _ = request("content_reviewer", "POST", "/v1/admin/books/"+bookID+"/approve", map[string]any{})
	check(200, status, nil)
	status, _ = request("school_admin", "POST", "/v1/admin/books/"+bookID+"/publish", map[string]any{})
	check(200, status, nil)
	status, outline := request("student", "GET", "/v1/books/"+bookID+"/outline", nil)
	check(200, status, outline)
	if len(outline["items"].([]any)) != 1 {
		t.Fatalf("outline: %+v", outline)
	}
	status, content := request("student", "GET", "/v1/lessons/"+lesson["id"].(string), nil)
	check(200, status, content)
	if content["content"] != "Synthetic educational content only." {
		t.Fatalf("lesson: %+v", content)
	}
	lessonID := lesson["id"].(string)
	status, emptyProgress := request("student", "GET", "/v1/books/"+bookID+"/progress", nil)
	check(200, status, emptyProgress)
	if len(emptyProgress["items"].([]any)) != 0 {
		t.Fatal("unexpected initial reading progress")
	}
	status, save := request("student", "PUT", "/v1/lessons/"+lessonID+"/progress", map[string]any{"completed": true, "bookmarked": true})
	check(200, status, save)
	if save["completed"] != true || save["bookmarked"] != true {
		t.Fatalf("reading progress not saved: %+v", save)
	}
	status, saved := request("student", "GET", "/v1/books/"+bookID+"/progress", nil)
	check(200, status, saved)
	if len(saved["items"].([]any)) != 1 {
		t.Fatalf("personal reading progress missing: %+v", saved)
	}
	status, _ = request("student", "PUT", "/v1/lessons/"+lessonID+"/progress", map[string]any{"completed": false, "bookmarked": true})
	check(200, status, nil)
	status, saved = request("student", "GET", "/v1/books/"+bookID+"/progress", nil)
	check(200, status, saved)
	record := saved["items"].([]any)[0].(map[string]any)
	if record["completed"] != false || record["bookmarked"] != true {
		t.Fatalf("reading status not updated: %+v", record)
	}
	status, _ = request("", "PUT", "/v1/lessons/"+lessonID+"/progress", map[string]any{"completed": true, "bookmarked": false})
	check(401, status, nil)
	status, _ = request("", "GET", "/v1/books/"+bookID+"/progress", nil)
	check(401, status, nil)
	var another string
	if err := db.QueryRowContext(ctx, "INSERT INTO users(email,display_name) VALUES($1,$2) RETURNING id::text", "other-"+suffix+"@example.test", "Other Reader").Scan(&another); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO school_memberships(school_id,user_id,role) VALUES($1::uuid,$2::uuid,'student')", school, another); err != nil {
		t.Fatal(err)
	}
	anotherToken, err := verifier.Issue(auth.Claims{Subject: another, SchoolID: school, Roles: []string{"student"}}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	tokens["other"] = anotherToken
	status, otherProgress := request("other", "GET", "/v1/books/"+bookID+"/progress", nil)
	check(200, status, otherProgress)
	if len(otherProgress["items"].([]any)) != 0 {
		t.Fatalf("another user's progress leaked: %+v", otherProgress)
	}
	var foreignSchool, foreignUser string
	if err := db.QueryRowContext(ctx, "INSERT INTO schools(name,slug) VALUES('Unrelated School',$1) RETURNING id::text", "foreign-"+suffix).Scan(&foreignSchool); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, "INSERT INTO users(email,display_name) VALUES($1,$2) RETURNING id::text", "foreign-"+suffix+"@example.test", "Foreign Reader").Scan(&foreignUser); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO school_memberships(school_id,user_id,role) VALUES($1::uuid,$2::uuid,'student')", foreignSchool, foreignUser); err != nil {
		t.Fatal(err)
	}
	tokens["foreign"], err = verifier.Issue(auth.Claims{Subject: foreignUser, SchoolID: foreignSchool, Roles: []string{"student"}}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	status, _ = request("foreign", "GET", "/v1/books/"+bookID+"/progress", nil)
	check(404, status, nil)
	status, _ = request("foreign", "PUT", "/v1/lessons/"+lessonID+"/progress", map[string]any{"completed": true, "bookmarked": true})
	check(404, status, nil)
	status, _ = request("", "GET", "/v1/lessons/"+lesson["id"].(string), nil)
	check(404, status, nil)
	status, _ = request("school_admin", "POST", "/v1/admin/books/"+bookID+"/chapters", map[string]any{"title": "Should fail", "position": 2})
	check(409, status, nil)
	status, _ = request("school_admin", "POST", "/v1/admin/books/"+bookID+"/archive", map[string]any{})
	check(200, status, nil)
	status, _ = request("student", "GET", "/v1/books/"+bookID+"/progress", nil)
	check(404, status, nil)
	status, _ = request("student", "PUT", "/v1/lessons/"+lessonID+"/progress", map[string]any{"completed": true, "bookmarked": true})
	check(404, status, nil)

}