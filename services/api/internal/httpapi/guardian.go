package httpapi

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"
)

func (s *Server) grantGuardian(w http.ResponseWriter, r *http.Request) {
	school := r.PathValue("school")
	_, ok := s.schoolScope(w, r, school, true)
	if !ok {
		return
	}
	var b struct {
		GuardianID string `json:"guardian_id"`
		StudentID  string `json:"student_id"`
	}
	if !decode(w, r, &b) {
		return
	}
	if !validID.MatchString(b.GuardianID) || !validID.MatchString(b.StudentID) || b.GuardianID == b.StudentID {
		fail(w, 400, "invalid relationship")
		return
	}
	var id string
	err := s.db.QueryRowContext(r.Context(), `INSERT INTO guardian_links(school_id,guardian_id,student_id,verified_at,revoked_at)
 SELECT $1::uuid,g.user_id,st.user_id,now(),NULL FROM school_memberships g
 JOIN school_memberships st ON st.school_id=g.school_id AND st.user_id=$3::uuid AND st.role='student' AND st.is_active=true
 WHERE g.school_id=$1::uuid AND g.user_id=$2::uuid AND g.role='guardian' AND g.is_active=true
 ON CONFLICT(school_id,guardian_id,student_id) DO UPDATE SET verified_at=now(),revoked_at=NULL RETURNING id::text`, school, b.GuardianID, b.StudentID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 409, "verified parent and student memberships required")
		return
	}
	if err != nil {
		fail(w, 503, "unable to verify guardian")
		return
	}
	respond(w, 201, map[string]string{"id": id, "status": "verified"})
}
func (s *Server) children(w http.ResponseWriter, r *http.Request) {
	c, ok, err := s.claims(r)
	if !ok || err != nil {
		fail(w, 401, "authentication required")
		return
	}
	if !c.HasRole("guardian") || c.SchoolID == "" {
		fail(w, 403, "verified guardian only")
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT DISTINCT gl.student_id::text,u.display_name,cs.class_id::text,cl.name FROM guardian_links gl
 JOIN users u ON u.id=gl.student_id AND u.status='active'
 LEFT JOIN class_students cs ON cs.school_id=gl.school_id AND cs.student_id=gl.student_id
 LEFT JOIN classes cl ON cl.school_id=cs.school_id AND cl.id=cs.class_id
 WHERE gl.school_id=$1::uuid AND gl.guardian_id=$2::uuid AND gl.verified_at IS NOT NULL AND gl.revoked_at IS NULL
 ORDER BY u.display_name LIMIT 100`, c.SchoolID, c.Subject)
	if err != nil {
		fail(w, 503, "unable to load children")
		return
	}
	defer rows.Close()
	items := []map[string]string{}
	for rows.Next() {
		var student, name string
		var classID, className sql.NullString
		if rows.Scan(&student, &name, &classID, &className) != nil {
			fail(w, 503, "unable to load children")
			return
		}
		items = append(items, map[string]string{"student_id": student, "name": name, "class_id": classID.String, "class_name": className.String})
	}
	if rows.Err() != nil {
		fail(w, 503, "unable to load children")
		return
	}
	respond(w, 200, map[string]any{"items": items})
}
func (s *Server) guardianChild(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	c, ok, err := s.claims(r)
	if !ok || err != nil {
		fail(w, 401, "authentication required")
		return "", "", false
	}
	if !c.HasRole("guardian") || c.SchoolID == "" {
		fail(w, 403, "guardian only")
		return "", "", false
	}
	student := r.PathValue("student")
	if !validID.MatchString(student) {
		fail(w, 400, "invalid student")
		return "", "", false
	}
	var verified bool
	err = s.db.QueryRowContext(r.Context(), `SELECT EXISTS(SELECT 1 FROM guardian_links WHERE school_id=$1::uuid AND guardian_id=$2::uuid AND student_id=$3::uuid AND verified_at IS NOT NULL AND revoked_at IS NULL)`, c.SchoolID, c.Subject, student).Scan(&verified)
	if err != nil || !verified {
		fail(w, 403, "guardian relationship not verified")
		return "", "", false
	}
	return c.SchoolID, c.Subject, true
}
func (s *Server) childAttendance(w http.ResponseWriter, r *http.Request) {
	school, _, ok := s.guardianChild(w, r)
	if !ok {
		return
	}
	student := r.PathValue("student")
	rows, err := s.db.QueryContext(r.Context(), `SELECT cl.name,ar.on_date::text,ar.status,ar.note FROM attendance_records ar
 JOIN classes cl ON cl.id=ar.class_id AND cl.school_id=ar.school_id
 WHERE ar.school_id=$1::uuid AND ar.student_id=$2::uuid AND ar.on_date >= current_date - interval '90 days'
 ORDER BY ar.on_date DESC LIMIT 100`, school, student)
	if err != nil {
		fail(w, 503, "unable to read attendance")
		return
	}
	defer rows.Close()
	items := []map[string]string{}
	for rows.Next() {
		var class, day, status, note string
		if rows.Scan(&class, &day, &status, &note) != nil {
			fail(w, 503, "unable to read attendance")
			return
		}
		items = append(items, map[string]string{"class_name": class, "date": day, "status": status, "note": note})
	}
	if rows.Err() != nil {
		fail(w, 503, "unable to read attendance")
		return
	}
	respond(w, 200, map[string]any{"items": items})
}
func (s *Server) childGrades(w http.ResponseWriter, r *http.Request) {
	school, _, ok := s.guardianChild(w, r)
	if !ok {
		return
	}
	student := r.PathValue("student")
	rows, err := s.db.QueryContext(r.Context(), `SELECT a.title,cl.name,sub.grade,COALESCE(sub.feedback,''),sub.submitted_at
 FROM assignment_submissions sub JOIN assignments a ON a.id=sub.assignment_id AND a.school_id=sub.school_id
 JOIN classes cl ON cl.id=a.class_id AND cl.school_id=a.school_id
 WHERE sub.school_id=$1::uuid AND sub.student_id=$2::uuid ORDER BY sub.submitted_at DESC LIMIT 100`, school, student)
	if err != nil {
		fail(w, 503, "unable to read grades")
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var title, class, feedback string
		var grade sql.NullFloat64
		var submitted time.Time
		if rows.Scan(&title, &class, &grade, &feedback, &submitted) != nil {
			fail(w, 503, "unable to read grades")
			return
		}
		var score any
		if grade.Valid {
			score = grade.Float64
		}
		items = append(items, map[string]any{"title": title, "class_name": class, "grade": score, "feedback": feedback, "submitted_at": submitted})
	}
	if rows.Err() != nil {
		fail(w, 503, "unable to read grades")
		return
	}
	// Only final submitted scores from the verified child's currently
	// authorized school are exposed. Neither question keys nor answers leave
	// the exam API through the family portal.
	quizRows, err := s.db.QueryContext(r.Context(), `SELECT q.title,cl.name,qa.grade,qa.submitted_at
 FROM quiz_attempts qa JOIN quizzes q ON q.id=qa.quiz_id AND q.school_id=qa.school_id
 JOIN classes cl ON cl.id=q.class_id AND cl.school_id=q.school_id
 WHERE qa.school_id=$1::uuid AND qa.student_id=$2::uuid
 AND qa.submitted_at IS NOT NULL AND qa.grade IS NOT NULL
 ORDER BY qa.submitted_at DESC LIMIT 100`, school, student)
	if err != nil {
		fail(w, 503, "unable to load exam grades")
		return
	}
	defer quizRows.Close()
	quizItems := []map[string]any{}
	for quizRows.Next() {
		var title, class string
		var grade float64
		var submitted time.Time
		if err := quizRows.Scan(&title, &class, &grade, &submitted); err != nil {
			fail(w, 503, "unable to load exam grades")
			return
		}
		quizItems = append(quizItems, map[string]any{
			"title": title, "class_name": class, "grade": grade,
			"submitted_at": submitted, "feedback": "",
		})
	}
	if quizRows.Err() != nil {
		fail(w, 503, "unable to load exam grades")
		return
	}
	respond(w, 200, map[string]any{"items": items, "quiz_items": quizItems})
}
func (s *Server) childLeave(w http.ResponseWriter, r *http.Request) {
	school, guardian, ok := s.guardianChild(w, r)
	if !ok {
		return
	}
	student := r.PathValue("student")
	var b struct {
		ClassID string `json:"class_id"`
		Start   string `json:"start_on"`
		End     string `json:"end_on"`
		Reason  string `json:"reason"`
	}
	if !decode(w, r, &b) {
		return
	}
	b.Reason = strings.TrimSpace(b.Reason)
	if !validID.MatchString(b.ClassID) || !isDay(b.Start) || !isDay(b.End) || b.End < b.Start || len(b.Reason) < 3 || len(b.Reason) > 1200 {
		fail(w, 400, "invalid leave request")
		return
	}
	var id string
	err := s.db.QueryRowContext(r.Context(), `INSERT INTO leave_requests(school_id,class_id,student_id,requested_by,start_on,end_on,reason)
 SELECT $1::uuid,cs.class_id,cs.student_id,$4::uuid,$5::date,$6::date,$7 FROM class_students cs
 WHERE cs.school_id=$1::uuid AND cs.student_id=$2::uuid AND cs.class_id=$3::uuid RETURNING id::text`, school, student, b.ClassID, guardian, b.Start, b.End, b.Reason).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 403, "child is not enrolled in this class")
		return
	}
	if err != nil {
		fail(w, 503, "unable to request leave")
		return
	}
	respond(w, 201, map[string]string{"id": id, "status": "pending"})
}
