package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/0xmarkhydra/OpenEdu/services/api/internal/auth"
)

func (s *Server) schoolScope(w http.ResponseWriter, r *http.Request, school string, write bool) (auth.Claims, bool) {
	c, ok, err := s.claims(r)
	if err != nil || !ok {
		fail(w, 401, "authentication required")
		return c, false
	}
	if !validID.MatchString(school) || c.SchoolID != school {
		fail(w, 403, "wrong school scope")
		return c, false
	}
	if write && !c.HasAny("school_admin", "principal") {
		fail(w, 403, "school administration required")
		return c, false
	}
	return c, true
}
func isDay(v string) bool {
	if len(v) != 10 {
		return false
	}
	date, err := time.Parse("2006-01-02", v)
	return err == nil && date.Format("2006-01-02") == v
}
func (s *Server) years(w http.ResponseWriter, r *http.Request) {
	school := r.PathValue("school")
	_, ok := s.schoolScope(w, r, school, r.Method == "POST")
	if !ok {
		return
	}
	if r.Method == "GET" {
		rows, err := s.db.QueryContext(r.Context(), `SELECT id::text,name,starts_on::text,ends_on::text FROM academic_years WHERE school_id=$1::uuid ORDER BY starts_on DESC LIMIT 50`, school)
		if err != nil {
			fail(w, 503, "unable to load academic years")
			return
		}
		defer rows.Close()
		items := []map[string]any{}
		for rows.Next() {
			var id, name, start, end string
			if rows.Scan(&id, &name, &start, &end) != nil {
				fail(w, 503, "unable to load academic years")
				return
			}
			items = append(items, map[string]any{"id": id, "name": name, "starts_on": start, "ends_on": end})
		}
		if rows.Err() != nil {
			fail(w, 503, "unable to load academic years")
			return
		}
		respond(w, 200, map[string]any{"items": items})
		return
	}
	var body struct {
		Name   string `json:"name"`
		Starts string `json:"starts_on"`
		Ends   string `json:"ends_on"`
	}
	if !decode(w, r, &body) {
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	if len(body.Name) < 3 || len(body.Name) > 80 || !isDay(body.Starts) || !isDay(body.Ends) || body.Ends <= body.Starts {
		fail(w, 400, "invalid academic year")
		return
	}
	var id string
	err := s.db.QueryRowContext(r.Context(), `INSERT INTO academic_years(school_id,name,starts_on,ends_on) VALUES($1::uuid,$2,$3::date,$4::date) ON CONFLICT(school_id,name) DO NOTHING RETURNING id::text`, school, body.Name, body.Starts, body.Ends).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 409, "academic year already exists")
		return
	}
	if err != nil {
		fail(w, 503, "unable to create academic year")
		return
	}
	respond(w, 201, map[string]any{"id": id, "name": body.Name, "starts_on": body.Starts, "ends_on": body.Ends})
}
func (s *Server) classes(w http.ResponseWriter, r *http.Request) {
	school := r.PathValue("school")
	c, ok := s.schoolScope(w, r, school, r.Method == "POST")
	if !ok {
		return
	}
	if r.Method == "GET" {
		rows, err := s.db.QueryContext(r.Context(), `SELECT c.id::text,c.academic_year_id::text,c.name,c.grade,ay.name FROM classes c JOIN academic_years ay ON ay.id=c.academic_year_id WHERE c.school_id=$1::uuid
 AND ($2::boolean
 OR EXISTS(SELECT 1 FROM class_students cs WHERE cs.school_id=c.school_id AND cs.class_id=c.id AND cs.student_id=$3::uuid)
 OR EXISTS(SELECT 1 FROM teacher_assignments ta WHERE ta.school_id=c.school_id AND ta.class_id=c.id AND ta.teacher_id=$3::uuid)
 OR EXISTS(SELECT 1 FROM guardian_links gl JOIN class_students cs ON cs.school_id=gl.school_id AND cs.student_id=gl.student_id
   WHERE gl.school_id=c.school_id AND cs.class_id=c.id AND gl.guardian_id=$3::uuid AND gl.verified_at IS NOT NULL AND gl.revoked_at IS NULL))
 ORDER BY ay.starts_on DESC,c.grade,c.name LIMIT 200`, school, c.HasAny("school_admin", "principal"), c.Subject)
		if err != nil {
			fail(w, 503, "unable to load classes")
			return
		}
		defer rows.Close()
		items := []map[string]any{}
		for rows.Next() {
			var id, year, name, yearName string
			var grade int
			if rows.Scan(&id, &year, &name, &grade, &yearName) != nil {
				fail(w, 503, "unable to load classes")
				return
			}
			items = append(items, map[string]any{"id": id, "academic_year_id": year, "name": name, "grade": grade, "academic_year": yearName})
		}
		if rows.Err() != nil {
			fail(w, 503, "unable to load classes")
			return
		}
		respond(w, 200, map[string]any{"items": items})
		return
	}
	var b struct {
		Name         string `json:"name"`
		AcademicYear string `json:"academic_year_id"`
		Grade        int    `json:"grade"`
	}
	if !decode(w, r, &b) {
		return
	}
	b.Name = strings.TrimSpace(b.Name)
	if len(b.Name) < 1 || len(b.Name) > 80 || !validID.MatchString(b.AcademicYear) || b.Grade < 1 || b.Grade > 12 {
		fail(w, 400, "invalid class")
		return
	}
	var id string
	err := s.db.QueryRowContext(r.Context(), `INSERT INTO classes(school_id,academic_year_id,name,grade)
 SELECT $1::uuid,ay.id,$3,$4 FROM academic_years ay WHERE ay.id=$2::uuid AND ay.school_id=$1::uuid
 ON CONFLICT(school_id,academic_year_id,name) DO NOTHING RETURNING id::text`, school, b.AcademicYear, b.Name, b.Grade).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 409, "year unavailable or class already exists")
		return
	}
	if err != nil {
		fail(w, 503, "unable to create class")
		return
	}
	respond(w, 201, map[string]any{"id": id, "name": b.Name, "academic_year_id": b.AcademicYear, "grade": b.Grade})
}
func (s *Server) classScope(w http.ResponseWriter, r *http.Request, id string, write bool) (auth.Claims, string, bool) {
	c, ok, err := s.claims(r)
	if err != nil || !ok {
		fail(w, 401, "authentication required")
		return c, "", false
	}
	if !validID.MatchString(id) {
		fail(w, 400, "invalid class id")
		return c, "", false
	}
	var school string
	err = s.db.QueryRowContext(r.Context(), `SELECT school_id::text FROM classes WHERE id=$1::uuid`, id).Scan(&school)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 404, "class not found")
		return c, "", false
	}
	if err != nil {
		fail(w, 503, "unable to check class")
		return c, "", false
	}
	if c.SchoolID != school {
		fail(w, 403, "wrong school scope")
		return c, "", false
	}
	if c.HasAny("school_admin", "principal") {
		return c, school, true
	}
	if c.HasRole("teacher") {
		var assigned bool
		err = s.db.QueryRowContext(r.Context(), `SELECT EXISTS(SELECT 1 FROM teacher_assignments WHERE school_id=$1::uuid AND class_id=$2::uuid AND teacher_id=$3::uuid)`, school, id, c.Subject).Scan(&assigned)
		if err == nil && assigned {
			return c, school, true
		}
	}
	if !write && c.HasRole("student") {
		var enrolled bool
		err = s.db.QueryRowContext(r.Context(), `SELECT EXISTS(SELECT 1 FROM class_students WHERE school_id=$1::uuid AND class_id=$2::uuid AND student_id=$3::uuid)`, school, id, c.Subject).Scan(&enrolled)
		if err == nil && enrolled {
			return c, school, true
		}
	}
	if !write && c.HasRole("guardian") {
		var verified bool
		err = s.db.QueryRowContext(r.Context(), `SELECT EXISTS(
		 SELECT 1 FROM guardian_links gl JOIN class_students cs ON cs.school_id=gl.school_id AND cs.student_id=gl.student_id
		 WHERE gl.school_id=$1::uuid AND cs.class_id=$2::uuid AND gl.guardian_id=$3::uuid
		 AND gl.verified_at IS NOT NULL AND gl.revoked_at IS NULL)`, school, id, c.Subject).Scan(&verified)
		if err == nil && verified {
			return c, school, true
		}
	}
	fail(w, 403, "class access forbidden")
	return c, "", false
}
func (s *Server) students(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, school, ok := s.classScope(w, r, id, r.Method == "POST")
	if !ok {
		return
	}
	if r.Method == "GET" {
		if !c.HasAny("school_admin", "principal", "teacher") {
			fail(w, 403, "student directory restricted")
			return
		}
		rows, err := s.db.QueryContext(r.Context(), `SELECT u.id::text,u.display_name FROM class_students cs JOIN users u ON u.id=cs.student_id WHERE cs.school_id=$1::uuid AND cs.class_id=$2::uuid ORDER BY u.display_name LIMIT 300`, school, id)
		if err != nil {
			fail(w, 503, "unable to load class roster")
			return
		}
		defer rows.Close()
		list := []map[string]string{}
		for rows.Next() {
			var uid, name string
			if rows.Scan(&uid, &name) != nil {
				fail(w, 503, "unable to load class roster")
				return
			}
			list = append(list, map[string]string{"id": uid, "name": name})
		}
		if rows.Err() != nil {
			fail(w, 503, "unable to load class roster")
			return
		}
		respond(w, 200, map[string]any{"items": list})
		return
	}
	if !c.HasAny("school_admin", "principal") {
		fail(w, 403, "school administration required")
		return
	}
	var b struct {
		StudentID string `json:"student_id"`
	}
	if !decode(w, r, &b) {
		return
	}
	if !validID.MatchString(b.StudentID) {
		fail(w, 400, "invalid student")
		return
	}
	result, err := s.db.ExecContext(r.Context(), `INSERT INTO class_students(school_id,class_id,student_id)
 SELECT $1::uuid,$2::uuid,m.user_id FROM school_memberships m
 WHERE m.school_id=$1::uuid AND m.user_id=$3::uuid AND m.role='student' AND m.is_active=true
 ON CONFLICT DO NOTHING`, school, id, b.StudentID)
	if err != nil {
		fail(w, 503, "unable to enroll student")
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		fail(w, 409, "student not eligible or already enrolled")
		return
	}
	respond(w, 201, map[string]any{"student_id": b.StudentID, "class_id": id})
}
func (s *Server) attendance(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, school, ok := s.classScope(w, r, id, true)
	if !ok {
		return
	}
	if r.Method == "GET" {
		day := r.URL.Query().Get("date")
		if !isDay(day) {
			fail(w, 400, "date must be YYYY-MM-DD")
			return
		}
		rows, err := s.db.QueryContext(r.Context(), `SELECT ar.student_id::text,u.display_name,ar.status,ar.note FROM attendance_records ar JOIN users u ON u.id=ar.student_id
 WHERE ar.school_id=$1::uuid AND ar.class_id=$2::uuid AND ar.on_date=$3::date ORDER BY u.display_name LIMIT 300`, school, id, day)
		if err != nil {
			fail(w, 503, "unable to load attendance")
			return
		}
		defer rows.Close()
		items := []map[string]string{}
		for rows.Next() {
			var uid, name, status, note string
			if rows.Scan(&uid, &name, &status, &note) != nil {
				fail(w, 503, "unable to load attendance")
				return
			}
			items = append(items, map[string]string{"student_id": uid, "name": name, "status": status, "note": note})
		}
		if rows.Err() != nil {
			fail(w, 503, "unable to load attendance")
			return
		}
		respond(w, 200, map[string]any{"items": items, "date": day})
		return
	}
	var b struct {
		StudentID string `json:"student_id"`
		Date      string `json:"date"`
		Status    string `json:"status"`
		Note      string `json:"note"`
	}
	if !decode(w, r, &b) {
		return
	}
	if !validID.MatchString(b.StudentID) || !isDay(b.Date) || len(b.Note) > 300 {
		fail(w, 400, "invalid attendance details")
		return
	}
	if b.Status != "present" && b.Status != "absent" && b.Status != "late" && b.Status != "excused" {
		fail(w, 400, "invalid status")
		return
	}
	res, err := s.db.ExecContext(r.Context(), `INSERT INTO attendance_records(school_id,class_id,student_id,on_date,status,note,recorded_by)
 SELECT $1::uuid,$2::uuid,cs.student_id,$4::date,$5,$6,$7::uuid FROM class_students cs WHERE cs.school_id=$1::uuid AND cs.class_id=$2::uuid AND cs.student_id=$3::uuid
 ON CONFLICT(school_id,class_id,student_id,on_date) DO UPDATE SET status=EXCLUDED.status,note=EXCLUDED.note,recorded_by=EXCLUDED.recorded_by,updated_at=now()`, school, id, b.StudentID, b.Date, b.Status, b.Note, c.Subject)
	if err != nil {
		fail(w, 503, "unable to record attendance")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		fail(w, 404, "student not enrolled")
		return
	}
	respond(w, 200, map[string]any{"student_id": b.StudentID, "date": b.Date, "status": b.Status})
}
func (s *Server) announcements(w http.ResponseWriter, r *http.Request) {
	school := r.PathValue("school")
	c, ok := s.schoolScope(w, r, school, false)
	if !ok {
		return
	}
	if r.Method == "GET" {
		audience := "all"
		if c.HasRole("teacher") {
			audience = "teacher"
		}
		if c.HasRole("student") {
			audience = "student"
		}
		if c.HasRole("guardian") {
			audience = "guardian"
		}
		if c.HasAny("school_admin", "principal", "content_admin") {
			audience = "*"
		}
		rows, err := s.db.QueryContext(r.Context(), `SELECT id::text,title,content,audience,created_at FROM school_announcements WHERE school_id=$1::uuid AND ($2='*' OR audience IN ('all',$2)) ORDER BY created_at DESC LIMIT 100`, school, audience)
		if err != nil {
			fail(w, 503, "unable to load announcements")
			return
		}
		defer rows.Close()
		items := []map[string]any{}
		for rows.Next() {
			var id, title, content, scope string
			var created time.Time
			if rows.Scan(&id, &title, &content, &scope, &created) != nil {
				fail(w, 503, "unable to load announcements")
				return
			}
			items = append(items, map[string]any{"id": id, "title": title, "content": content, "audience": scope, "created_at": created})
		}
		if rows.Err() != nil {
			fail(w, 503, "unable to load announcements")
			return
		}
		respond(w, 200, map[string]any{"items": items})
		return
	}
	if !c.HasAny("school_admin", "principal") {
		fail(w, 403, "school administration required")
		return
	}
	var b struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		Audience string `json:"audience"`
	}
	if !decode(w, r, &b) {
		return
	}
	b.Title = strings.TrimSpace(b.Title)
	b.Content = strings.TrimSpace(b.Content)
	if len(b.Title) < 3 || len(b.Title) > 180 || len(b.Content) < 3 || len(b.Content) > 5000 {
		fail(w, 400, "invalid announcement")
		return
	}
	if b.Audience != "all" && b.Audience != "teacher" && b.Audience != "student" && b.Audience != "guardian" {
		fail(w, 400, "invalid audience")
		return
	}
	var id string
	err := s.db.QueryRowContext(r.Context(), `INSERT INTO school_announcements(school_id,title,content,audience,created_by) VALUES($1::uuid,$2,$3,$4,$5::uuid) RETURNING id::text`, school, b.Title, b.Content, b.Audience, c.Subject).Scan(&id)
	if err != nil {
		fail(w, 503, "unable to publish announcement")
		return
	}
	respond(w, 201, map[string]any{"id": id, "title": b.Title, "audience": b.Audience})
}
func (s *Server) overview(w http.ResponseWriter, r *http.Request) {
	school := r.PathValue("school")
	_, ok := s.schoolScope(w, r, school, false)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	var classes, students, teachers, announcements int
	err := s.db.QueryRowContext(ctx, `SELECT
 (SELECT count(*) FROM classes WHERE school_id=$1::uuid),
 (SELECT count(DISTINCT student_id) FROM class_students WHERE school_id=$1::uuid),
 (SELECT count(DISTINCT teacher_id) FROM teacher_assignments WHERE school_id=$1::uuid),
 (SELECT count(*) FROM school_announcements WHERE school_id=$1::uuid)`, school).Scan(&classes, &students, &teachers, &announcements)
	if err != nil {
		fail(w, 503, "unable to calculate overview")
		return
	}
	respond(w, 200, map[string]any{"classes": classes, "students": students, "teachers": teachers, "announcements": announcements})
}
func (s *Server) assignTeacher(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, school, ok := s.classScope(w, r, id, true)
	if !ok {
		return
	}
	if !c.HasAny("school_admin", "principal") {
		fail(w, 403, "school administration required")
		return
	}
	var b struct {
		TeacherID string `json:"teacher_id"`
		Subject   string `json:"subject_code"`
	}
	if !decode(w, r, &b) {
		return
	}
	b.Subject = strings.TrimSpace(b.Subject)
	if !validID.MatchString(b.TeacherID) || len(b.Subject) < 2 || len(b.Subject) > 48 {
		fail(w, 400, "invalid teacher assignment")
		return
	}
	res, err := s.db.ExecContext(r.Context(), `INSERT INTO teacher_assignments(school_id,class_id,teacher_id,subject_code)
 SELECT $1::uuid,$2::uuid,m.user_id,$4 FROM school_memberships m
 WHERE m.school_id=$1::uuid AND m.user_id=$3::uuid AND m.role='teacher' AND m.is_active=true
 ON CONFLICT DO NOTHING`, school, id, b.TeacherID, b.Subject)
	if err != nil {
		fail(w, 503, "unable to assign teacher")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		fail(w, 409, "teacher not eligible or already assigned")
		return
	}
	respond(w, 201, map[string]any{"class_id": id, "teacher_id": b.TeacherID, "subject_code": b.Subject})
}
func (s *Server) timetable(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, school, ok := s.classScope(w, r, id, r.Method == "POST")
	if !ok {
		return
	}
	if r.Method == "GET" {
		rows, err := s.db.QueryContext(r.Context(), `SELECT id::text,weekday,period,subject_code,COALESCE(teacher_id::text,''),room FROM timetable_entries WHERE school_id=$1::uuid AND class_id=$2::uuid ORDER BY weekday,period LIMIT 105`, school, id)
		if err != nil {
			fail(w, 503, "unable to load timetable")
			return
		}
		defer rows.Close()
		items := []map[string]any{}
		for rows.Next() {
			var uid, subject, teacher, room string
			var day, period int
			if rows.Scan(&uid, &day, &period, &subject, &teacher, &room) != nil {
				fail(w, 503, "unable to load timetable")
				return
			}
			items = append(items, map[string]any{"id": uid, "weekday": day, "period": period, "subject_code": subject, "teacher_id": teacher, "room": room})
		}
		if rows.Err() != nil {
			fail(w, 503, "unable to load timetable")
			return
		}
		respond(w, 200, map[string]any{"items": items})
		return
	}
	if !c.HasAny("school_admin", "principal") {
		fail(w, 403, "school administration required")
		return
	}
	var b struct {
		Weekday int    `json:"weekday"`
		Period  int    `json:"period"`
		Subject string `json:"subject_code"`
		Room    string `json:"room"`
	}
	if !decode(w, r, &b) {
		return
	}
	b.Subject = strings.TrimSpace(b.Subject)
	if b.Weekday < 1 || b.Weekday > 7 || b.Period < 1 || b.Period > 15 || len(b.Subject) < 2 || len(b.Subject) > 60 || len(b.Room) > 50 {
		fail(w, 400, "invalid timetable entry")
		return
	}
	var uid string
	err := s.db.QueryRowContext(r.Context(), `INSERT INTO timetable_entries(school_id,class_id,weekday,period,subject_code,room) VALUES($1::uuid,$2::uuid,$3,$4,$5,$6)
 ON CONFLICT(class_id,weekday,period) DO UPDATE SET subject_code=EXCLUDED.subject_code,room=EXCLUDED.room RETURNING id::text`, school, id, b.Weekday, b.Period, b.Subject, b.Room).Scan(&uid)
	if err != nil {
		fail(w, 503, "unable to update timetable")
		return
	}
	respond(w, 200, map[string]any{"id": uid, "weekday": b.Weekday, "period": b.Period, "subject_code": b.Subject, "room": b.Room})
}

// Shared tenant scope must be validated at every route; do not infer tenant from headers.
