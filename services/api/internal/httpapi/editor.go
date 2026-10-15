package httpapi

import (
	"database/sql"
	"net/http"
	"time"
)

func (s *Server) adminBooks(w http.ResponseWriter, r *http.Request) {
	c, ok, err := s.claims(r)
	if err != nil || !ok {
		fail(w, 401, "authentication required")
		return
	}
	if !c.HasAny("platform_admin", "school_admin", "content_admin", "content_reviewer") {
		fail(w, 403, "editor role required")
		return
	}
	if c.SchoolID == "" && !c.HasRole("platform_admin") {
		fail(w, 403, "invalid platform scope")
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT id::text,title,grade,subject_code,status,visibility,created_at
 FROM books WHERE ($1='' AND school_id IS NULL) OR ($1<>'' AND school_id=NULLIF($1,'')::uuid)
 ORDER BY updated_at DESC LIMIT 100`, c.SchoolID)
	if err != nil {
		fail(w, 503, "unable to load editorial catalog")
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, title, subject, status, visibility string
		var grade int
		var created time.Time
		if rows.Scan(&id, &title, &grade, &subject, &status, &visibility, &created) != nil {
			fail(w, 503, "unable to load editorial catalog")
			return
		}
		items = append(items, map[string]any{"id": id, "title": title, "grade": grade, "subject_code": subject, "status": status, "visibility": visibility, "created_at": created})
	}
	if rows.Err() != nil {
		fail(w, 503, "unable to load editorial catalog")
		return
	}
	respond(w, 200, map[string]any{"items": items})
}

func (s *Server) draftOutline(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, ok := s.authorBook(w, r, id)
	if !ok {
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT c.id::text,c.title,c.position,l.id::text,l.title,l.position
 FROM chapters c LEFT JOIN lessons l ON l.chapter_id=c.id WHERE c.book_id=$1::uuid
 ORDER BY c.position,l.position LIMIT 300`, id)
	if err != nil {
		fail(w, 503, "unable to read draft chapters")
		return
	}
	defer rows.Close()
	type l struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Position int    `json:"position"`
	}
	type ch struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Position int    `json:"position"`
		Lessons  []l    `json:"lessons"`
	}
	list := []ch{}
	prev := ""
	for rows.Next() {
		var id, title string
		var pos int
		var lessonID, lessonTitle sql.NullString
		var lessonPos sql.NullInt64
		if rows.Scan(&id, &title, &pos, &lessonID, &lessonTitle, &lessonPos) != nil {
			fail(w, 503, "unable to read draft chapters")
			return
		}
		if prev != id {
			list = append(list, ch{ID: id, Title: title, Position: pos, Lessons: []l{}})
			prev = id
		}
		if lessonID.Valid {
			list[len(list)-1].Lessons = append(list[len(list)-1].Lessons, l{ID: lessonID.String, Title: lessonTitle.String, Position: int(lessonPos.Int64)})
		}
	}
	if rows.Err() != nil {
		fail(w, 503, "unable to read draft chapters")
		return
	}
	respond(w, 200, map[string]any{"items": list})
}
