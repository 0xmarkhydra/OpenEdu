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

	// Server-graded tests cannot leak answer keys or allow outsiders to start.
	due := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
	status, _ = request("student", "POST", classRoute+"/quizzes", map[string]any{"title": "Forbidden quiz", "duration_minutes": 30, "due_at": due})
	check(403, status, nil)
	status, quiz := request("teacher", "POST", classRoute+"/quizzes", map[string]any{"title": "Safe math review", "duration_minutes": 30, "due_at": due})
	check(201, status, quiz)
	qid := quiz["id"].(string)
	status, _ = request("student", "GET", "/v1/quizzes/"+qid, nil)
	check(404, status, nil)
	status, _ = request("teacher", "POST", "/v1/quizzes/"+qid+"/publish", map[string]any{})
	check(409, status, nil)
	status, q1 := request("teacher", "POST", "/v1/quizzes/"+qid+"/questions", map[string]any{
		"prompt": "Which answer equals 4?", "options": []string{"1", "2", "3", "4"}, "correct_option": 3, "position": 1})
	check(201, status, q1)
	status, q2 := request("teacher", "POST", "/v1/quizzes/"+qid+"/questions", map[string]any{
		"prompt": "Which answer equals 9?", "options": []string{"6", "7", "9", "8"}, "correct_option": 2, "position": 2})
	check(201, status, q2)
	status, preview := request("teacher", "GET", "/v1/quizzes/"+qid, nil)
	check(200, status, preview)
	if preview["items"].([]any)[0].(map[string]any)["correct_option"] == nil {
		t.Fatalf("teacher preview missing key: %+v", preview)
	}
	status, _ = request("teacher", "POST", "/v1/quizzes/"+qid+"/publish", map[string]any{})
	check(200, status, nil)
	status, _ = request("teacher", "POST", "/v1/quizzes/"+qid+"/questions", map[string]any{
		"prompt": "Late edit", "options": []string{"1", "2", "3", "4"}, "correct_option": 0, "position": 3})
	check(409, status, nil)
	status, _ = request("guardian", "GET", "/v1/quizzes/"+qid, nil)
	check(403, status, nil)

	// A valid JWT issued for another school cannot enumerate this quiz.
	var outsideID string
	if err := db.QueryRowContext(ctx, "INSERT INTO users(email,display_name) VALUES($1,'Outsider') RETURNING id::text", "outside-"+suffix+"@example.test").Scan(&outsideID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO school_memberships(school_id,user_id,role) VALUES($1::uuid,$2::uuid,'student')", schoolB, outsideID); err != nil {
		t.Fatal(err)
	}
	tokens["outsider"], err = v.Issue(auth.Claims{Subject: outsideID, SchoolID: schoolB, Roles: []string{"student"}}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	status, _ = request("outsider", "GET", "/v1/quizzes/"+qid, nil)
	check(403, status, nil)
	status, _ = request("outsider", "POST", "/v1/quizzes/"+qid+"/start", map[string]any{})
	check(403, status, nil)

	status, _ = request("principal", "GET", "/v1/quizzes/"+qid+"/results", nil)
	check(200, status, nil)
	status, notStarted := request("student", "GET", "/v1/quizzes/"+qid, nil)
	check(200, status, notStarted)
	if len(notStarted["items"].([]any)) != 0 {
		t.Fatal("exam questions leaked before start")
	}
	status, _ = request("student", "POST", "/v1/quizzes/"+qid+"/submit", map[string]any{"answers": map[string]int{q1["id"].(string): 3, q2["id"].(string): 2}})
	check(409, status, nil)
	status, started := request("student", "POST", "/v1/quizzes/"+qid+"/start", map[string]any{})
	check(200, status, started)
	status, startAgain := request("student", "POST", "/v1/quizzes/"+qid+"/start", map[string]any{})
	check(200, status, startAgain)
	if started["expires_at"] != startAgain["expires_at"] {
		t.Fatal("attempt time reset by start")
	}
	status, exam := request("student", "GET", "/v1/quizzes/"+qid, nil)
	check(200, status, exam)
	if len(exam["items"].([]any)) != 2 {
		t.Fatal("student missing questions")
	}
	for _, item := range exam["items"].([]any) {
		if _, leak := item.(map[string]any)["correct_option"]; leak {
			t.Fatal("student leaked grading key")
		}
	}
	status, _ = request("student", "POST", "/v1/quizzes/"+qid+"/submit", map[string]any{
		"answers": map[string]int{q1["id"].(string): 3}})
	check(400, status, nil)
	status, grade := request("student", "POST", "/v1/quizzes/"+qid+"/submit", map[string]any{
		"answers": map[string]int{q1["id"].(string): 3, q2["id"].(string): 0}})
	check(200, status, grade)
	if grade["grade"].(float64) != 5 {
		t.Fatalf("incorrect auto-grade: %+v", grade)
	}
	status, _ = request("student", "POST", "/v1/quizzes/"+qid+"/submit", map[string]any{
		"answers": map[string]int{q1["id"].(string): 3, q2["id"].(string): 2}})
	check(409, status, nil)
	status, afterSubmit := request("student", "GET", "/v1/quizzes/"+qid, nil)
	check(200, status, afterSubmit)
	if len(afterSubmit["items"].([]any)) != 0 || afterSubmit["grade"].(float64) != 5 {
		t.Fatalf("exam leaked after submit: %+v", afterSubmit)
	}
	status, results := request("teacher", "GET", "/v1/quizzes/"+qid+"/results", nil)
	check(200, status, results)
	if len(results["items"].([]any)) != 1 {
		t.Fatalf("quiz results missing: %+v", results)
	}
	status, _ = request("student", "GET", "/v1/quizzes/"+qid+"/results", nil)
	check(403, status, nil)
	status, _ = request("principal", "GET", "/v1/classes/"+schoolB+"/quizzes", nil)
	check(404, status, nil)

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

	status, unverifiedContacts := request("guardian", "GET", "/v1/me/messages/contacts", nil)
	check(200, status, unverifiedContacts)
	if len(unverifiedContacts["items"].([]any)) != 0 {
		t.Fatalf("contact leaked before verification: %+v", unverifiedContacts)
	}
	status, unauthorizedMessage := request("guardian", "POST", "/v1/me/messages", map[string]any{"class_id": classID, "recipient_id": ids["teacher"], "content": "Unverified guardian message"})
	check(403, status, unauthorizedMessage)
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
	status, verifiedContacts := request("guardian", "GET", "/v1/me/messages/contacts", nil)
	check(200, status, verifiedContacts)
	if len(verifiedContacts["items"].([]any)) != 1 {
		t.Fatalf("missing teacher contact: %+v", verifiedContacts)
	}
	status, sent := request("guardian", "POST", "/v1/me/messages", map[string]any{"class_id": classID, "recipient_id": ids["teacher"], "content": "Con xin được tham gia ôn tập"})
	check(201, status, sent)
	status, inbox := request("teacher", "GET", "/v1/me/messages", nil)
	check(200, status, inbox)
	if len(inbox["items"].([]any)) != 1 {
		t.Fatalf("missing message: %+v", inbox)
	}
	messageID := sent["id"].(string)
	status, msg := request("teacher", "PATCH", "/v1/me/messages/"+messageID+"/read", map[string]any{})
	check(200, status, msg)
	status, unread := request("student", "GET", "/v1/me/messages", nil)
	check(403, status, unread)
	status, msg = request("guardian", "PATCH", "/v1/me/messages/"+messageID+"/read", map[string]any{})
	check(404, status, msg)
	status, msg = request("guardian", "POST", "/v1/me/messages", map[string]any{"class_id": classID, "recipient_id": ids["principal"], "content": "Not an assigned teacher"})
	check(403, status, msg)
	status, parentLeave := request("guardian", "POST", "/v1/me/children/"+ids["student"]+"/leave", map[string]any{"class_id": classID, "start_on": "2026-10-20", "end_on": "2026-10-21", "reason": "Xin nghỉ có phép"})
	check(201, status, parentLeave)
	status, _ = request("guardian", "POST", path+"/guardians", map[string]any{"guardian_id": ids["guardian"], "student_id": ids["student"]})
	check(403, status, nil)
	if _, err := db.ExecContext(ctx, `UPDATE guardian_links SET revoked_at=now() WHERE school_id=$1::uuid AND guardian_id=$2::uuid AND student_id=$3::uuid`, schoolA, ids["guardian"], ids["student"]); err != nil {
		t.Fatal(err)
	}
	status, blocked := request("guardian", "GET", "/v1/me/messages", nil)
	check(200, status, blocked)
	if len(blocked["items"].([]any)) != 0 {
		t.Fatalf("revoked guardian retained private messages: %+v", blocked)
	}
	status, _ = request("guardian", "POST", "/v1/me/messages", map[string]any{"class_id": classID, "recipient_id": ids["teacher"], "content": "Revoked request"})
	check(403, status, nil)
	status, report := request("principal", "GET", path+"/reports", nil)
	check(200, status, report)
	stats := report["summary"].(map[string]any)
	if stats["total_students"].(float64) != 1 || stats["assignments"].(float64) != 1 {
		t.Fatalf("bad report: %+v", stats)
	}
	status, _ = request("student", "GET", path+"/reports", nil)
	check(403, status, nil)
	status, _ = request("principal", "GET", "/v1/schools/"+schoolB+"/reports", nil)
	check(403, status, nil)
	status, overview := request("principal", "GET", path+"/overview", nil)
	check(200, status, overview)
	if overview["classes"].(float64) != 1 || overview["students"].(float64) != 1 || overview["teachers"].(float64) != 1 {
		t.Fatalf("wrong overview: %+v", overview)
	}

	// Server deadline cannot be bypassed by a submitted answer after expiry.
	var lateStudent string
	if err := db.QueryRowContext(ctx, "INSERT INTO users(email,display_name) VALUES($1,'Late Student') RETURNING id::text", "late-"+suffix+"@example.test").Scan(&lateStudent); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO school_memberships(school_id,user_id,role) VALUES($1::uuid,$2::uuid,'student')", schoolA, lateStudent); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO class_students(school_id,class_id,student_id) VALUES($1::uuid,$2::uuid,$3::uuid)", schoolA, classID, lateStudent); err != nil {
		t.Fatal(err)
	}
	tokens["late"], err = v.Issue(auth.Claims{Subject: lateStudent, SchoolID: schoolA, Roles: []string{"student"}}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	status, _ = request("late", "POST", "/v1/quizzes/"+qid+"/start", map[string]any{})
	check(200, status, nil)
	if _, err := db.ExecContext(ctx, "UPDATE quiz_attempts SET expires_at=now()-interval '1 minute' WHERE quiz_id=$1::uuid AND student_id=$2::uuid", qid, lateStudent); err != nil {
		t.Fatal(err)
	}
	status, _ = request("late", "POST", "/v1/quizzes/"+qid+"/submit", map[string]any{"answers": map[string]int{q1["id"].(string): 3, q2["id"].(string): 2}})
	check(409, status, nil)
	status, expired := request("late", "GET", "/v1/quizzes/"+qid, nil)
	check(200, status, expired)
	if len(expired["items"].([]any)) != 0 {
		t.Fatal("expired attempt leaked exam questions")
	}

}
