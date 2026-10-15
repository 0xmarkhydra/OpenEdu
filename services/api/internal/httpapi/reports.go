package httpapi

import (
	"net/http"
)

// Aggregate-only reports, tenant-scoped and limited to school leadership.
// No student names, messages, personal notes or raw attendance records.
func (s *Server) schoolReport(w http.ResponseWriter, r *http.Request) {
	school := r.PathValue("school")
	_, ok := s.schoolScope(w, r, school, true)
	if !ok {
		return
	}
	stats := struct {
		TotalStudents     int `json:"total_students"`
		TotalClasses      int `json:"total_classes"`
		AttendancePresent int `json:"attendance_present_30d"`
		AttendanceTotal   int `json:"attendance_total_30d"`
		Assignments       int `json:"assignments"`
		Submissions       int `json:"submissions"`
		Graded            int `json:"graded"`
		PendingLeave      int `json:"pending_leave"`
	}{}
	err := s.db.QueryRowContext(r.Context(), `SELECT
 (SELECT count(DISTINCT student_id) FROM class_students WHERE school_id=$1::uuid),
 (SELECT count(*) FROM classes WHERE school_id=$1::uuid),
 (SELECT count(*) FROM attendance_records WHERE school_id=$1::uuid AND on_date>=current_date-interval '30 days' AND status IN ('present','late')),
 (SELECT count(*) FROM attendance_records WHERE school_id=$1::uuid AND on_date>=current_date-interval '30 days'),
 (SELECT count(*) FROM assignments WHERE school_id=$1::uuid),
 (SELECT count(*) FROM assignment_submissions WHERE school_id=$1::uuid),
 (SELECT count(*) FROM assignment_submissions WHERE school_id=$1::uuid AND grade IS NOT NULL),
 (SELECT count(*) FROM leave_requests WHERE school_id=$1::uuid AND status='pending')`, school).
		Scan(&stats.TotalStudents, &stats.TotalClasses, &stats.AttendancePresent, &stats.AttendanceTotal, &stats.Assignments, &stats.Submissions, &stats.Graded, &stats.PendingLeave)
	if err != nil {
		fail(w, 503, "unable to load report")
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT c.id::text,c.name,c.grade,count(cs.student_id)
 FROM classes c LEFT JOIN class_students cs ON cs.school_id=c.school_id AND cs.class_id=c.id
 WHERE c.school_id=$1::uuid GROUP BY c.id,c.name,c.grade ORDER BY c.grade,c.name LIMIT 150`, school)
	if err != nil {
		fail(w, 503, "unable to load class summary")
		return
	}
	defer rows.Close()
	classes := []map[string]any{}
	for rows.Next() {
		var id, name string
		var grade, students int
		if rows.Scan(&id, &name, &grade, &students) != nil {
			fail(w, 503, "unable to load class summary")
			return
		}
		classes = append(classes, map[string]any{"id": id, "name": name, "grade": grade, "students": students})
	}
	if rows.Err() != nil {
		fail(w, 503, "unable to load class summary")
		return
	}
	respond(w, 200, map[string]any{
		"summary": stats,
		"classes": classes,
		"period":  "rolling_30_days",
	})
}
