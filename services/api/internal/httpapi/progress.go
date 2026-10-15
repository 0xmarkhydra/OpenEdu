package httpapi

import (
	"database/sql"
	"errors"
	"net/http"
)

// Personal reading state is intentionally not part of guardian or analytics API.
// Access to the underlying published material is checked on every read/write.
func (s *Server) bookProgress(w http.ResponseWriter, r *http.Request) {
	c, ok, err := s.claims(r)
	if err != nil || !ok {
		fail(w, 401, "authentication required")
		return
	}
	bookID := r.PathValue("id")
	if _, visible := s.visibleBook(w, r, bookID); !visible {
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT rp.lesson_id::text,rp.is_completed,rp.is_bookmarked,rp.last_opened_at
 FROM reader_progress rp
 JOIN lessons l ON l.id=rp.lesson_id
 JOIN chapters ch ON ch.id=l.chapter_id
 WHERE rp.user_id=$1::uuid AND ch.book_id=$2::uuid
 ORDER BY rp.last_opened_at DESC LIMIT 500`, c.Subject, bookID)
	if err != nil {
		fail(w, 503, "unable to load progress")
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var lessonID string
		var completed, bookmarked bool
		var when sql.NullTime
		if err := rows.Scan(&lessonID, &completed, &bookmarked, &when); err != nil {
			fail(w, 503, "unable to load progress")
			return
		}
		items = append(items, map[string]any{"lesson_id": lessonID, "completed": completed, "bookmarked": bookmarked, "last_opened_at": when.Time})
	}
	if rows.Err() != nil {
		fail(w, 503, "unable to load progress")
		return
	}
	respond(w, 200, map[string]any{"items": items})
}
func (s *Server) lessonProgress(w http.ResponseWriter, r *http.Request) {
	c, exists, err := s.claims(r)
	if err != nil || !exists {
		fail(w, 401, "authentication required")
		return
	}
	lessonID := r.PathValue("id")
	if !validID.MatchString(lessonID) {
		fail(w, 400, "invalid lesson")
		return
	}
	var b struct {
		Completed  bool `json:"completed"`
		Bookmarked bool `json:"bookmarked"`
	}
	if !decode(w, r, &b) {
		return
	}
	// INSERT ... SELECT is atomic with the visibility predicate. Tokens only carry
	// a school ID after live principal validation (s.claims), preventing spoofing.
	var savedLesson string
	var completed, bookmarked bool
	err = s.db.QueryRowContext(r.Context(), `INSERT INTO reader_progress(user_id,lesson_id,is_completed,is_bookmarked)
 SELECT $1::uuid,l.id,$4::boolean,$5::boolean FROM lessons l
 JOIN chapters ch ON ch.id=l.chapter_id
 JOIN books b ON b.id=ch.book_id
 WHERE l.id=$2::uuid AND b.status='published' AND
 ((b.visibility='public' AND b.school_id IS NULL) OR
  (b.visibility='school' AND b.school_id=NULLIF($3,'')::uuid))
 ON CONFLICT(user_id,lesson_id) DO UPDATE SET
 is_completed=EXCLUDED.is_completed,is_bookmarked=EXCLUDED.is_bookmarked,
 last_opened_at=now(),updated_at=now()
 RETURNING lesson_id::text,is_completed,is_bookmarked`, c.Subject, lessonID, c.SchoolID, b.Completed, b.Bookmarked).Scan(&savedLesson, &completed, &bookmarked)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 404, "lesson not published or unavailable")
		return
	}
	if err != nil {
		fail(w, 503, "unable to save progress")
		return
	}
	respond(w, 200, map[string]any{"lesson_id": savedLesson, "completed": completed, "bookmarked": bookmarked})
}
