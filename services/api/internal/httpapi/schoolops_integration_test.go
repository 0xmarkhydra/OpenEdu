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

// This test is never allowed to access production or a non-disposable database.
func TestSchoolLearningOperationsPostgres(t *testing.T) {
	dsn := os.Getenv("OPENEDU_TEST_DSN")
	if dsn == "" {
		t.Skip("set OPENEDU_TEST_DSN for disposable integration test")
	}
	if !strings.Contains(dsn, "openedu_test") || strings.Contains(dsn, "railway") {
		t.Fatal("test DB guard")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Second)
	defer cancel()
	if err := migrations.Apply(ctx, db); err != nil {
		t.Fatal("migrate: ", err)
	}
	suffix := fmt.Sprintf("ops-%d", time.Now().UnixNano())
	var schoolA, schoolB string
	for j, p := range []*string{&schoolA, &schoolB} {
		if err := db.QueryRowContext(ctx, "INSERT INTO schools(name,slug) VALUES($1,$2) RETURNING id::text", "Operations Test", fmt.Sprintf("%s-%d", suffix, j)).Scan(p); err != nil {
			t.Fatal(err)
		}
	}
	users := []string{"principal", "teacher", "student", "guardian"}
	ids := map[string]string{}
	for _, role := range users {
		var id string
		if err := db.QueryRowContext(ctx, "INSERT INTO users(email,display_name) VALUES($1,$2) RETURNING id::text", role+suffix+"@example.test", role).Scan(&id); err != nil {
			t.Fatal(err)
		}
		ids[role] = id
		if _, err = db.ExecContext(ctx, "INSERT INTO school_memberships(school_id,user_id,role) VALUES($1::uuid,$2::uuid,$3)", schoolA, id, role); err != nil {
			t.Fatal(err)
		}
	}
	v, err := auth.NewVerifier("01234567890123456789012345678901", "openedu-test", "openedu-api")
	if err != nil {
		t.Fatal(err)
	}
	tokens := map[string]string{}
	for _, role := range users {
		token, err := v.Issue(auth.Claims{Subject: ids[role], SchoolID: schoolA, Roles: []string{role}}, time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		tokens[role] = token
	}
	app := httptest.NewServer(New(catalog.NewPostgresStore(db), v, db))
	defer app.Close()
	request := func(role, method, path string, input any) (int, map[string]any) {
		var body []byte
		if input != nil {
			body, _ = json.Marshal(input)
		}
		req, err := http.NewRequestWithContext(ctx, method, app.URL+path, bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+tokens[role])
		req.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var out map[string]any
		if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
			t.Fatal(err)
		}
		return res.StatusCode, out
	}
	check := func(expected, actual int, data map[string]any) {
		if actual != expected {
			t.Fatalf("want HTTP %d got %d: %+v", expected, actual, data)
		}
	}
	path := "/v1/schools/" + schoolA
	status, year := request("principal", "POST", path+"/years", map[string]any{"name": "2026-2027", "starts_on": "2026-09-01", "ends_on": "2027-06-30"})
	check(201, status, year)
	status, _ = request("principal", "GET", "/v1/schools/"+schoolB+"/years", nil)
	check(403, status, nil)
	status, _ = request("student", "POST", path+"/classes", map[string]any{"name": "6A", "academic_year_id": year["id"], "grade": 6})
	check(403, status, nil)
	status, cls := request("principal", "POST", path+"/classes", map[string]any{"name": "6A", "academic_year_id": year["id"], "grade": 6})
	check(201, status, cls)
	classID := cls["id"].(string)
	classRoute := "/v1/classes/" + classID
	status, _ = request("principal", "POST", classRoute+"/students", map[string]any{"student_id": ids["student"]})
	check(201, status, nil)
	status, _ = request("principal", "POST", classRoute+"/teachers", map[string]any{"teacher_id": ids["teacher"], "subject_code": "toan"})
	check(201, status, nil)
	status, classes := request("student", "GET", path+"/classes", nil)
	check(200, status, classes)
	if len(classes["items"].([]any)) != 1 {
		t.Fatalf("student sees wrong classes: %+v", classes)
	}
	status, notice := request("principal", "POST", path+"/announcements", map[string]any{"title": "Thông báo họp", "content": "Nhà trường có thông báo", "audience": "all"})
	check(201, status, notice)
	status, board := request("student", "GET", path+"/announcements", nil)
	check(200, status, board)
	if len(board["items"].([]any)) == 0 {
		t.Fatal("student cannot read announcement")
	}
	status, slot := request("principal", "POST", classRoute+"/timetable", map[string]any{"weekday": 2, "period": 3, "subject_code": "toan", "room": "B101"})
	check(200, status, slot)
	status, _ = request("principal", "GET", classRoute+"/timetable", nil)
	check(200, status, nil)
	status, _ = request("teacher", "PUT", classRoute+"/attendance", map[string]any{"student_id": ids["student"], "date": "2026-10-15", "status": "present", "note": ""})
	check(200, status, nil)
	status, att := request("teacher", "GET", classRoute+"/attendance?date=2026-10-15", nil)
	check(200, status, att)
	if len(att["items"].([]any)) != 1 {
		t.Fatal("attendance missing")
	}
	status, hw := request("teacher", "POST", classRoute+"/assignments", map[string]any{"title": "Bài ôn tập chương 1", "instructions": "Trả lời các câu hỏi"})
	check(201, status, hw)
	aid := hw["id"].(string)
	status, _ = request("student", "POST", "/v1/assignments/"+aid+"/submissions", map[string]any{"content": "Bài làm của học sinh"})
	check(200, status, nil)
	status, _ = request("teacher", "PATCH", "/v1/assignments/"+aid+"/submissions", map[string]any{"student_id": ids["student"], "grade": 9, "feedback": "Làm tốt"})
	check(200, status, nil)
	status, subs := request("student", "GET", "/v1/assignments/"+aid+"/submissions", nil)
	check(200, status, subs)
	if len(subs["items"].([]any)) != 1 {
		t.Fatal("student submission unavailable")
	}
	status, leave := request("student", "POST", classRoute+"/leave-requests", map[string]any{"start_on": "2026-10-16", "end_on": "2026-10-17", "reason": "Gia đình có việc"})
	check(201, status, leave)
	status, _ = request("teacher", "PATCH", classRoute+"/leave-requests", map[string]any{"request_id": leave["id"], "status": "approved"})
	check(200, status, nil)

	status, denied := request("guardian", "GET", "/v1/me/children/"+ids["student"]+"/grades", nil)
	check(403, status, denied)
	status, relation := request("principal", "POST", path+"/guardians", map[string]any{"guardian_id": ids["guardian"], "student_id": ids["student"]})
	check(201, status, relation)
	status, children := request("guardian", "GET", "/v1/me/children", nil)
	check(200, status, children)
	if len(children["items"].([]any)) != 1 {
		t.Fatalf("missing verified child: %+v", children)
	}
	status, records := request("guardian", "GET", "/v1/me/children/"+ids["student"]+"/grades", nil)
	check(200, status, records)
	if len(records["items"].([]any)) != 1 {
		t.Fatalf("unexpected guardian grade records: %+v", records)
	}
	status, attendanceForParent := request("guardian", "GET", "/v1/me/children/"+ids["student"]+"/attendance", nil)
	check(200, status, attendanceForParent)
	if attendanceForParent["items"] == nil {
		t.Fatalf("missing attendance payload: %+v", attendanceForParent)
	}
	status, parentLeave := request("guardian", "POST", "/v1/me/children/"+ids["student"]+"/leave", map[string]any{"class_id": classID, "start_on": "2026-10-20", "end_on": "2026-10-21", "reason": "Xin nghỉ có phép"})
	check(201, status, parentLeave)
	status, _ = request("guardian", "POST", path+"/guardians", map[string]any{"guardian_id": ids["guardian"], "student_id": ids["student"]})
	check(403, status, nil)
	status, overview := request("principal", "GET", path+"/overview", nil)
	check(200, status, overview)
	if overview["classes"].(float64) != 1 || overview["students"].(float64) != 1 || overview["teachers"].(float64) != 1 {
		t.Fatalf("wrong overview: %+v", overview)
	}
}
