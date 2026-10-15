package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/0xmarkhydra/OpenEdu/services/api/internal/auth"
)

// Quizzes are scoped to a currently enrolled student or a currently assigned
// teacher/school administrator. The client never receives the answer key.
func (s *Server) quizScope(w http.ResponseWriter, r *http.Request) (auth.Claims, string, string, string, int, time.Time, bool, bool) {
	id := r.PathValue("id")
	if !validID.MatchString(id) {
		fail(w, 400, "invalid quiz")
		return auth.Claims{}, "", "", "", 0, time.Time{}, false, false
	}
	var school, class, status string
	var duration int
	var due time.Time
	err := s.db.QueryRowContext(r.Context(), `SELECT school_id::text,class_id::text,status,duration_minutes,due_at FROM quizzes WHERE id=$1::uuid`, id).Scan(&school, &class, &status, &duration, &due)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 404, "quiz unavailable")
		return auth.Claims{}, "", "", "", 0, time.Time{}, false, false
	}
	if err != nil {
		fail(w, 503, "unable to load quiz")
		return auth.Claims{}, "", "", "", 0, time.Time{}, false, false
	}
	c, scope, ok := s.classScope(w, r, class, false)
	if !ok {
		return c, "", "", "", 0, time.Time{}, false, false
	}
	if scope != school {
		fail(w, 403, "wrong school")
		return c, "", "", "", 0, time.Time{}, false, false
	}
	staff := c.HasAny("principal", "school_admin", "teacher")
	if !staff && !c.HasRole("student") {
		fail(w, 403, "students or teachers only")
		return c, "", "", "", 0, time.Time{}, false, false
	}
	if !staff && status != "published" {
		fail(w, 404, "quiz not published")
		return c, "", "", "", 0, time.Time{}, false, false
	}
	return c, school, class, status, duration, due, staff, true
}
func (s *Server) classQuizzes(w http.ResponseWriter, r *http.Request) {
	class := r.PathValue("id")
	c, school, ok := s.classScope(w, r, class, r.Method == "POST")
	if !ok {
		return
	}
	staff := c.HasAny("principal", "school_admin", "teacher")
	if !staff && !c.HasRole("student") {
		fail(w, 403, "students or teachers only")
		return
	}
	if r.Method == "GET" {
		rows, err := s.db.QueryContext(r.Context(), `SELECT q.id::text,q.title,q.duration_minutes,q.due_at,q.status,
  (SELECT COUNT(*) FROM quiz_questions WHERE quiz_id=q.id),
  (SELECT qa.grade FROM quiz_attempts qa WHERE qa.quiz_id=q.id AND qa.student_id=$3::uuid),
  (SELECT qa.submitted_at FROM quiz_attempts qa WHERE qa.quiz_id=q.id AND qa.student_id=$3::uuid)
  FROM quizzes q WHERE q.school_id=$1::uuid AND q.class_id=$2::uuid
  AND ($4::boolean OR q.status='published')
  ORDER BY q.created_at DESC LIMIT 100`, school, class, c.Subject, staff)
		if err != nil {
			fail(w, 503, "unable to load quizzes")
			return
		}
		defer rows.Close()
		items := []map[string]any{}
		for rows.Next() {
			var id, title, status string
			var duration, count int
			var due time.Time
			var grade sql.NullFloat64
			var submitted sql.NullTime
			if err := rows.Scan(&id, &title, &duration, &due, &status, &count, &grade, &submitted); err != nil {
				fail(w, 503, "unable to load quizzes")
				return
			}
			var score any
			if grade.Valid {
				score = grade.Float64
			}
			items = append(items, map[string]any{"id": id, "title": title, "duration_minutes": duration, "due_at": due, "status": status, "questions": count, "grade": score, "submitted": submitted.Valid})
		}
		if rows.Err() != nil {
			fail(w, 503, "unable to load quizzes")
			return
		}
		respond(w, 200, map[string]any{"items": items})
		return
	}
	if !staff {
		fail(w, 403, "staff required")
		return
	}
	var body struct {
		Title    string `json:"title"`
		Duration int    `json:"duration_minutes"`
		Due      string `json:"due_at"`
	}
	if !decode(w, r, &body) {
		return
	}
	body.Title = strings.TrimSpace(body.Title)
	due, err := time.Parse(time.RFC3339, body.Due)
	if err != nil || len(body.Title) < 3 || len(body.Title) > 180 || body.Duration < 5 || body.Duration > 180 || !due.After(time.Now()) {
		fail(w, 400, "invalid quiz")
		return
	}
	var id string
	err = s.db.QueryRowContext(r.Context(), `INSERT INTO quizzes(school_id,class_id,title,duration_minutes,due_at,created_by)
 VALUES($1::uuid,$2::uuid,$3,$4,$5,$6::uuid) RETURNING id::text`, school, class, body.Title, body.Duration, due, c.Subject).Scan(&id)
	if err != nil {
		fail(w, 503, "unable to create quiz")
		return
	}
	respond(w, 201, map[string]any{"id": id, "status": "draft"})
}
func (s *Server) quizQuestions(w http.ResponseWriter, r *http.Request) {
	c, _, _, status, _, _, staff, ok := s.quizScope(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if !staff {
		fail(w, 403, "teacher required")
		return
	}
	if status != "draft" {
		fail(w, 409, "published questions are immutable")
		return
	}
	var b struct {
		Prompt   string   `json:"prompt"`
		Options  []string `json:"options"`
		Correct  int      `json:"correct_option"`
		Position int      `json:"position"`
	}
	if !decode(w, r, &b) {
		return
	}
	b.Prompt = strings.TrimSpace(b.Prompt)
	if len(b.Prompt) < 3 || len(b.Prompt) > 2000 || len(b.Options) != 4 || b.Correct < 0 || b.Correct > 3 || b.Position < 1 || b.Position > 100 {
		fail(w, 400, "invalid question")
		return
	}
	for i, v := range b.Options {
		b.Options[i] = strings.TrimSpace(v)
		if len(b.Options[i]) < 1 || len(b.Options[i]) > 500 {
			fail(w, 400, "invalid options")
			return
		}
	}
	var created string
	err := s.db.QueryRowContext(r.Context(), `INSERT INTO quiz_questions(quiz_id,position,prompt,option_a,option_b,option_c,option_d,correct_option)
 SELECT q.id,$2,$3,$4,$5,$6,$7,$8 FROM quizzes q WHERE q.id=$1::uuid AND q.status='draft' AND
 EXISTS(SELECT 1 FROM school_memberships m WHERE m.user_id=$9::uuid AND m.school_id=q.school_id AND m.is_active=true)
 ON CONFLICT(quiz_id,position) DO NOTHING RETURNING id::text`, id, b.Position, b.Prompt, b.Options[0], b.Options[1], b.Options[2], b.Options[3], b.Correct, c.Subject).Scan(&created)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 409, "quiz closed or question position exists")
		return
	}
	if err != nil {
		fail(w, 503, "unable to add question")
		return
	}
	respond(w, 201, map[string]any{"id": created, "position": b.Position})
}
func (s *Server) publishQuiz(w http.ResponseWriter, r *http.Request) {
	_, _, _, status, _, due, staff, ok := s.quizScope(w, r)
	if !ok {
		return
	}
	if !staff {
		fail(w, 403, "staff required")
		return
	}
	if status != "draft" {
		fail(w, 409, "quiz already published")
		return
	}
	if !due.After(time.Now()) {
		fail(w, 409, "quiz already expired")
		return
	}
	result, err := s.db.ExecContext(r.Context(), `UPDATE quizzes q SET status='published'
 WHERE q.id=$1::uuid AND q.status='draft' AND q.due_at>now()
 AND EXISTS(SELECT 1 FROM quiz_questions t WHERE t.quiz_id=q.id)`, r.PathValue("id"))
	if err != nil {
		fail(w, 503, "unable to publish quiz")
		return
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		fail(w, 409, "add questions before publishing")
		return
	}
	respond(w, 200, map[string]string{"status": "published"})
}
func (s *Server) quizDetail(w http.ResponseWriter, r *http.Request) {
	c, _, _, status, _, due, staff, ok := s.quizScope(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	var started, expires, submitted sql.NullTime
	var grade sql.NullFloat64
	if !staff {
		err := s.db.QueryRowContext(r.Context(), `SELECT started_at,expires_at,submitted_at,grade FROM quiz_attempts WHERE quiz_id=$1::uuid AND student_id=$2::uuid`, id, c.Subject).Scan(&started, &expires, &submitted, &grade)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			fail(w, 503, "unable to load attempt")
			return
		}
	}
	var score any
	if grade.Valid {
		score = grade.Float64
	}
	var deadline any
	if expires.Valid {
		deadline = expires.Time
	}
	response := map[string]any{"id": id, "status": status, "due_at": due, "started": started.Valid, "expires_at": deadline, "submitted": submitted.Valid, "grade": score, "items": []any{}}
	// Neither keys nor questions can be requested by students outside a live attempt.
	if !staff && (!started.Valid || submitted.Valid || !time.Now().Before(expires.Time) || !time.Now().Before(due)) {
		respond(w, 200, response)
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT id::text,prompt,position,option_a,option_b,option_c,option_d,correct_option
 FROM quiz_questions WHERE quiz_id=$1::uuid ORDER BY position LIMIT 100`, id)
	if err != nil {
		fail(w, 503, "unable to load questions")
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var qid, prompt, a, b, d, e string
		var pos, correct int
		if rows.Scan(&qid, &prompt, &pos, &a, &b, &d, &e, &correct) != nil {
			fail(w, 503, "unable to load questions")
			return
		}
		item := map[string]any{"id": qid, "position": pos, "prompt": prompt, "options": []string{a, b, d, e}}
		if staff {
			item["correct_option"] = correct
		}
		items = append(items, item)
	}
	if rows.Err() != nil {
		fail(w, 503, "unable to load questions")
		return
	}
	response["items"] = items
	respond(w, 200, response)
}
func (s *Server) startQuiz(w http.ResponseWriter, r *http.Request) {
	c, school, _, status, duration, due, staff, ok := s.quizScope(w, r)
	if !ok {
		return
	}
	if staff || !c.HasRole("student") {
		fail(w, 403, "student required")
		return
	}
	if status != "published" || !due.After(time.Now()) {
		fail(w, 409, "quiz unavailable")
		return
	}
	var expiry time.Time
	var submitted sql.NullTime
	var grade sql.NullFloat64
	err := s.db.QueryRowContext(r.Context(), `INSERT INTO quiz_attempts(school_id,quiz_id,student_id,expires_at)
 SELECT $1::uuid,q.id,$3::uuid,LEAST(q.due_at,now()+($4::int*interval '1 minute'))
 FROM quizzes q WHERE q.id=$2::uuid AND q.school_id=$1::uuid AND q.status='published'
 AND q.due_at>now() AND EXISTS(SELECT 1 FROM class_students cs
 WHERE cs.school_id=q.school_id AND cs.class_id=q.class_id AND cs.student_id=$3::uuid)
 ON CONFLICT(quiz_id,student_id) DO UPDATE SET student_id=EXCLUDED.student_id
 RETURNING expires_at,submitted_at,grade`, school, r.PathValue("id"), c.Subject, duration).Scan(&expiry, &submitted, &grade)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 403, "not enrolled or quiz closed")
		return
	}
	if err != nil {
		fail(w, 503, "unable to begin quiz")
		return
	}
	var score any
	if grade.Valid {
		score = grade.Float64
	}
	respond(w, 200, map[string]any{"expires_at": expiry, "submitted": submitted.Valid, "grade": score})
}
func (s *Server) submitQuiz(w http.ResponseWriter, r *http.Request) {
	c, school, _, status, _, due, staff, ok := s.quizScope(w, r)
	if !ok {
		return
	}
	if staff || !c.HasRole("student") {
		fail(w, 403, "student required")
		return
	}
	if status != "published" {
		fail(w, 409, "quiz closed")
		return
	}
	var body struct {
		Answers map[string]int `json:"answers"`
	}
	if !decode(w, r, &body) {
		return
	}
	if len(body.Answers) < 1 || len(body.Answers) > 100 {
		fail(w, 400, "answers required")
		return
	}
	for id, option := range body.Answers {
		if !validID.MatchString(id) || option < 0 || option > 3 {
			fail(w, 400, "invalid answer")
			return
		}
	}
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		fail(w, 503, "unable to begin submission")
		return
	}
	defer tx.Rollback()
	var expires time.Time
	var submitted sql.NullTime
	err = tx.QueryRowContext(r.Context(), `SELECT expires_at,submitted_at FROM quiz_attempts
 WHERE school_id=$1::uuid AND quiz_id=$2::uuid AND student_id=$3::uuid FOR UPDATE`, school, r.PathValue("id"), c.Subject).Scan(&expires, &submitted)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 409, "start quiz before submitting")
		return
	}
	if err != nil {
		fail(w, 503, "unable to load attempt")
		return
	}
	if submitted.Valid || !time.Now().Before(expires) || !time.Now().Before(due) {
		fail(w, 409, "quiz already submitted or expired")
		return
	}
	rows, err := tx.QueryContext(r.Context(), `SELECT id::text,correct_option FROM quiz_questions WHERE quiz_id=$1::uuid`, r.PathValue("id"))
	if err != nil {
		fail(w, 503, "unable to load grading keys")
		return
	}
	count, correct := 0, 0
	for rows.Next() {
		var id string
		var answer int
		if rows.Scan(&id, &answer) != nil {
			rows.Close()
			fail(w, 503, "unable to grade")
			return
		}
		choice, found := body.Answers[id]
		if !found {
			rows.Close()
			fail(w, 400, "answer every question")
			return
		}
		if choice == answer {
			correct++
		}
		count++
	}
	e := rows.Err()
	rows.Close()
	if e != nil {
		fail(w, 503, "unable to grade")
		return
	}
	if count == 0 || len(body.Answers) != count {
		fail(w, 400, "invalid question list")
		return
	}
	score := math.Round((float64(correct) * 1000 / float64(count))) / 100
	payload, _ := json.Marshal(body.Answers)
	result, err := tx.ExecContext(r.Context(), `UPDATE quiz_attempts SET answers=$4::jsonb,grade=$5,submitted_at=now()
 WHERE school_id=$1::uuid AND quiz_id=$2::uuid AND student_id=$3::uuid
 AND submitted_at IS NULL AND expires_at>now()`, school, r.PathValue("id"), c.Subject, string(payload), score)
	if err != nil {
		fail(w, 503, "unable to save submission")
		return
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		fail(w, 409, "quiz expired or already submitted")
		return
	}
	if err = tx.Commit(); err != nil {
		fail(w, 503, "unable to commit submission")
		return
	}
	respond(w, 200, map[string]any{"status": "graded", "grade": score, "correct": correct, "total": count})
}
func (s *Server) quizResults(w http.ResponseWriter, r *http.Request) {
	_, school, _, _, _, _, staff, ok := s.quizScope(w, r)
	if !ok {
		return
	}
	if !staff {
		fail(w, 403, "staff required")
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT qa.student_id::text,u.display_name,qa.started_at,qa.submitted_at,qa.grade
 FROM quiz_attempts qa JOIN users u ON u.id=qa.student_id
 WHERE qa.school_id=$1::uuid AND qa.quiz_id=$2::uuid ORDER BY qa.started_at DESC LIMIT 200`, school, r.PathValue("id"))
	if err != nil {
		fail(w, 503, "unable to load results")
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, name string
		var start time.Time
		var submitted sql.NullTime
		var grade sql.NullFloat64
		if rows.Scan(&id, &name, &start, &submitted, &grade) != nil {
			fail(w, 503, "unable to load results")
			return
		}
		var score any
		if grade.Valid {
			score = grade.Float64
		}
		items = append(items, map[string]any{"student_id": id, "name": name, "started_at": start, "submitted": submitted.Valid, "grade": score})
	}
	if rows.Err() != nil {
		fail(w, 503, "unable to load results")
		return
	}
	respond(w, 200, map[string]any{"items": items})
}
