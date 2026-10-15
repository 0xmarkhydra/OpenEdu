package httpapi

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"
)

func (s *Server) assignments(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, school, ok := s.classScope(w, r, id, r.Method == "POST")
	if !ok {
		return
	}
	if r.Method == "GET" {
		rows, err := s.db.QueryContext(r.Context(), `SELECT id::text,title,instructions,due_at,created_at FROM assignments WHERE school_id=$1::uuid AND class_id=$2::uuid ORDER BY created_at DESC LIMIT 100`, school, id)
		if err != nil {
			fail(w, 503, "unable to load assignments")
			return
		}
		defer rows.Close()
		items := []map[string]any{}
		for rows.Next() {
			var aid, title, instructions string
			var due sql.NullTime
			var created time.Time
			if rows.Scan(&aid, &title, &instructions, &due, &created) != nil {
				fail(w, 503, "unable to load assignments")
				return
			}
			var deadline any
			if due.Valid {
				deadline = due.Time
			}
			items = append(items, map[string]any{"id": aid, "title": title, "instructions": instructions, "due_at": deadline, "created_at": created})
		}
		if rows.Err() != nil {
			fail(w, 503, "unable to load assignments")
			return
		}
		respond(w, 200, map[string]any{"items": items})
		return
	}
	if !c.HasAny("teacher", "school_admin", "principal") {
		fail(w, 403, "teachers only")
		return
	}
	var b struct {
		Title        string `json:"title"`
		Instructions string `json:"instructions"`
		Due          string `json:"due_at"`
	}
	if !decode(w, r, &b) {
		return
	}
	b.Title = strings.TrimSpace(b.Title)
	if len(b.Title) < 3 || len(b.Title) > 180 || len(b.Instructions) > 10000 {
		fail(w, 400, "invalid assignment")
		return
	}
	var due any
	if b.Due != "" {
		t, err := time.Parse(time.RFC3339, b.Due)
		if err != nil {
			fail(w, 400, "invalid deadline")
			return
		}
		due = t
	}
	var aid string
	err := s.db.QueryRowContext(r.Context(), `INSERT INTO assignments(school_id,class_id,title,instructions,due_at,created_by) VALUES($1::uuid,$2::uuid,$3,$4,$5,$6::uuid) RETURNING id::text`, school, id, b.Title, b.Instructions, due, c.Subject).Scan(&aid)
	if err != nil {
		fail(w, 503, "unable to create assignment")
		return
	}
	respond(w, 201, map[string]any{"id": aid, "title": b.Title, "instructions": b.Instructions, "due_at": due})
}
func (s *Server) submissionScope(w http.ResponseWriter, r *http.Request, assignmentID string) (string, string, bool) {
	c, ok, err := s.claims(r)
	if err != nil || !ok {
		fail(w, 401, "authentication required")
		return "", "", false
	}
	if !validID.MatchString(assignmentID) {
		fail(w, 400, "invalid assignment")
		return "", "", false
	}
	var school, class string
	err = s.db.QueryRowContext(r.Context(), `SELECT school_id::text,class_id::text FROM assignments WHERE id=$1::uuid`, assignmentID).Scan(&school, &class)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 404, "assignment not found")
		return "", "", false
	}
	if err != nil {
		fail(w, 503, "unable to load assignment")
		return "", "", false
	}
	if c.SchoolID != school {
		fail(w, 403, "school access forbidden")
		return "", "", false
	}
	return school, class, true
}
func (s *Server) submissions(w http.ResponseWriter, r *http.Request) {
	aid := r.PathValue("id")
	school, class, ok := s.submissionScope(w, r, aid)
	if !ok {
		return
	}
	c, _, _ := s.claims(r)
	if r.Method == "POST" {
		if !c.HasRole("student") {
			fail(w, 403, "students only")
			return
		}
		var enrolled bool
		if err := s.db.QueryRowContext(r.Context(), `SELECT EXISTS(SELECT 1 FROM class_students WHERE school_id=$1::uuid AND class_id=$2::uuid AND student_id=$3::uuid)`, school, class, c.Subject).Scan(&enrolled); err != nil || !enrolled {
			fail(w, 403, "not enrolled")
			return
		}
		var b struct {
			Content string `json:"content"`
		}
		if !decode(w, r, &b) {
			return
		}
		b.Content = strings.TrimSpace(b.Content)
		if len(b.Content) < 1 || len(b.Content) > 20000 {
			fail(w, 400, "invalid submission")
			return
		}
		var submitted time.Time
		err := s.db.QueryRowContext(r.Context(), `INSERT INTO assignment_submissions(school_id,assignment_id,student_id,content) VALUES($1::uuid,$2::uuid,$3::uuid,$4)
  ON CONFLICT(school_id,assignment_id,student_id) DO UPDATE SET content=EXCLUDED.content,submitted_at=now(),grade=NULL,feedback=NULL,graded_at=NULL
  RETURNING submitted_at`, school, aid, c.Subject, b.Content).Scan(&submitted)
		if err != nil {
			fail(w, 503, "unable to submit work")
			return
		}
		respond(w, 200, map[string]any{"status": "submitted", "submitted_at": submitted})
		return
	}
	isStaff := c.HasAny("school_admin", "principal")
	if c.HasRole("teacher") {
		var assigned bool
		err := s.db.QueryRowContext(r.Context(), `SELECT EXISTS(SELECT 1 FROM teacher_assignments WHERE school_id=$1::uuid AND class_id=$2::uuid AND teacher_id=$3::uuid)`, school, class, c.Subject).Scan(&assigned)
		if err == nil && assigned {
			isStaff = true
		}
	}
	if r.Method == "PATCH" {
		if !isStaff {
			fail(w, 403, "teacher review required")
			return
		}
		var b struct {
			StudentID string  `json:"student_id"`
			Grade     float64 `json:"grade"`
			Feedback  string  `json:"feedback"`
		}
		if !decode(w, r, &b) {
			return
		}
		if !validID.MatchString(b.StudentID) || b.Grade < 0 || b.Grade > 10 || len(b.Feedback) > 5000 {
			fail(w, 400, "invalid grade")
			return
		}
		res, err := s.db.ExecContext(r.Context(), `UPDATE assignment_submissions SET grade=$4,feedback=$5,graded_at=now() WHERE school_id=$1::uuid AND assignment_id=$2::uuid AND student_id=$3::uuid`, school, aid, b.StudentID, b.Grade, b.Feedback)
		if err != nil {
			fail(w, 503, "unable to grade")
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			fail(w, 404, "submission not found")
			return
		}
		respond(w, 200, map[string]any{"status": "graded", "student_id": b.StudentID, "grade": b.Grade})
		return
	}
	if !isStaff && !c.HasRole("student") {
		fail(w, 403, "restricted submissions")
		return
	}
	var student any
	if !isStaff {
		student = c.Subject
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT sub.student_id::text,u.display_name,sub.content,sub.grade,sub.feedback,sub.submitted_at FROM assignment_submissions sub JOIN users u ON u.id=sub.student_id
 WHERE sub.school_id=$1::uuid AND sub.assignment_id=$2::uuid AND ($3::uuid IS NULL OR sub.student_id=$3::uuid) ORDER BY sub.submitted_at DESC LIMIT 200`, school, aid, student)
	if err != nil {
		fail(w, 503, "unable to load submissions")
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var uid, name, content string
		var grade sql.NullFloat64
		var feedback sql.NullString
		var submitted time.Time
		if rows.Scan(&uid, &name, &content, &grade, &feedback, &submitted) != nil {
			fail(w, 503, "unable to load submissions")
			return
		}
		var score any
		if grade.Valid {
			score = grade.Float64
		}
		items = append(items, map[string]any{"student_id": uid, "name": name, "content": content, "grade": score, "feedback": feedback.String, "submitted_at": submitted})
	}
	if rows.Err() != nil {
		fail(w, 503, "unable to load submissions")
		return
	}
	respond(w, 200, map[string]any{"items": items})
}
func (s *Server) leaveRequests(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, school, ok := s.classScope(w, r, id, false)
	if !ok {
		return
	}
	if r.Method == "GET" {
		if !c.HasAny("teacher", "principal", "school_admin", "student") {
			fail(w, 403, "leave requests restricted")
			return
		}
		studentParam := any(nil)
		if c.HasRole("student") && !c.HasAny("school_admin", "principal", "teacher") {
			studentParam = c.Subject
		}
		rows, err := s.db.QueryContext(r.Context(), `SELECT id::text,student_id::text,start_on::text,end_on::text,reason,status,created_at FROM leave_requests
  WHERE school_id=$1::uuid AND class_id=$2::uuid AND ($3::uuid IS NULL OR student_id=$3::uuid) ORDER BY created_at DESC LIMIT 100`, school, id, studentParam)
		if err != nil {
			fail(w, 503, "unable to load leave requests")
			return
		}
		defer rows.Close()
		items := []map[string]any{}
		for rows.Next() {
			var uid, student, start, end, reason, status string
			var created time.Time
			if rows.Scan(&uid, &student, &start, &end, &reason, &status, &created) != nil {
				fail(w, 503, "unable to load leave requests")
				return
			}
			items = append(items, map[string]any{"id": uid, "student_id": student, "start_on": start, "end_on": end, "reason": reason, "status": status, "created_at": created})
		}
		if rows.Err() != nil {
			fail(w, 503, "unable to load leave requests")
			return
		}
		respond(w, 200, map[string]any{"items": items})
		return
	}
	if r.Method == "POST" {
		if !c.HasRole("student") {
			fail(w, 403, "student request required")
			return
		}
		var b struct {
			Start  string `json:"start_on"`
			End    string `json:"end_on"`
			Reason string `json:"reason"`
		}
		if !decode(w, r, &b) {
			return
		}
		b.Reason = strings.TrimSpace(b.Reason)
		if !isDay(b.Start) || !isDay(b.End) || b.End < b.Start || len(b.Reason) < 3 || len(b.Reason) > 1200 {
			fail(w, 400, "invalid leave request")
			return
		}
		var uid string
		err := s.db.QueryRowContext(r.Context(), `INSERT INTO leave_requests(school_id,class_id,student_id,requested_by,start_on,end_on,reason)
    SELECT $1::uuid,$2::uuid,cs.student_id,$3::uuid,$4::date,$5::date,$6 FROM class_students cs
    WHERE cs.school_id=$1::uuid AND cs.class_id=$2::uuid AND cs.student_id=$3::uuid RETURNING id::text`, school, id, c.Subject, b.Start, b.End, b.Reason).Scan(&uid)
		if errors.Is(err, sql.ErrNoRows) {
			fail(w, 403, "not enrolled")
			return
		}
		if err != nil {
			fail(w, 503, "unable to create leave request")
			return
		}
		respond(w, 201, map[string]any{"id": uid, "status": "pending"})
		return
	}
	if !c.HasAny("teacher", "principal", "school_admin") {
		fail(w, 403, "review permission required")
		return
	}
	var b struct {
		RequestID string `json:"request_id"`
		Status    string `json:"status"`
	}
	if !decode(w, r, &b) {
		return
	}
	if !validID.MatchString(b.RequestID) || (b.Status != "approved" && b.Status != "rejected") {
		fail(w, 400, "invalid decision")
		return
	}
	res, err := s.db.ExecContext(r.Context(), `UPDATE leave_requests SET status=$4,resolved_by=$5::uuid,resolved_at=now()
  WHERE id=$3::uuid AND school_id=$1::uuid AND class_id=$2::uuid AND status='pending'`, school, id, b.RequestID, b.Status, c.Subject)
	if err != nil {
		fail(w, 503, "unable to resolve request")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		fail(w, 409, "request unavailable or already reviewed")
		return
	}
	respond(w, 200, map[string]any{"id": b.RequestID, "status": b.Status})
}
