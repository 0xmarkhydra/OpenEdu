package httpapi

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/0xmarkhydra/OpenEdu/services/api/internal/auth"
	"github.com/0xmarkhydra/OpenEdu/services/api/internal/catalog"
)

func (s *Server) visibleBook(w http.ResponseWriter, r *http.Request, id string) (catalog.Book, bool) {
	if !validID.MatchString(id) {
		fail(w, 400, "invalid book")
		return catalog.Book{}, false
	}
	c, _, err := s.claims(r)
	if err != nil {
		fail(w, 401, "invalid token")
		return catalog.Book{}, false
	}
	book, err := s.store.Get(r.Context(), c.SchoolID, id)
	if errors.Is(err, catalog.ErrNotFound) {
		fail(w, 404, "book unavailable")
		return catalog.Book{}, false
	}
	if err != nil {
		fail(w, 503, "unable to load book")
		return catalog.Book{}, false
	}
	return book, true
}
func (s *Server) outline(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, ok := s.visibleBook(w, r, id)
	if !ok {
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT c.id::text,c.title,c.position,l.id::text,l.title,l.position,l.content_type FROM chapters c LEFT JOIN lessons l ON l.chapter_id=c.id WHERE c.book_id=$1::uuid ORDER BY c.position,l.position LIMIT 300`, id)
	if err != nil {
		fail(w, 503, "unable to load outline")
		return
	}
	defer rows.Close()
	type lesson struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Position int    `json:"position"`
		Kind     string `json:"content_type"`
	}
	type chapter struct {
		ID       string   `json:"id"`
		Title    string   `json:"title"`
		Position int      `json:"position"`
		Lessons  []lesson `json:"lessons"`
	}
	chapters := []chapter{}
	last := ""
	for rows.Next() {
		var cid, title string
		var position int
		var lid, ltitle, kind sql.NullString
		var lposition sql.NullInt64
		if err := rows.Scan(&cid, &title, &position, &lid, &ltitle, &lposition, &kind); err != nil {
			fail(w, 503, "unable to load outline")
			return
		}
		if cid != last {
			chapters = append(chapters, chapter{ID: cid, Title: title, Position: position, Lessons: []lesson{}})
			last = cid
		}
		if lid.Valid {
			chapters[len(chapters)-1].Lessons = append(chapters[len(chapters)-1].Lessons, lesson{ID: lid.String, Title: ltitle.String, Position: int(lposition.Int64), Kind: kind.String})
		}
	}
	if rows.Err() != nil {
		fail(w, 503, "unable to load outline")
		return
	}
	respond(w, 200, map[string]any{"items": chapters})
}
func (s *Server) lessonContent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID.MatchString(id) {
		fail(w, 400, "invalid lesson")
		return
	}
	c, _, err := s.claims(r)
	if err != nil {
		fail(w, 401, "invalid token")
		return
	}
	var title, kind, content, bookID string
	err = s.db.QueryRowContext(r.Context(), `SELECT l.title,l.content_type,COALESCE(l.content_body,''),b.id::text
 FROM lessons l JOIN chapters ch ON ch.id=l.chapter_id JOIN books b ON b.id=ch.book_id
 WHERE l.id=$1::uuid AND b.status='published'
 AND ((b.visibility='public' AND b.school_id IS NULL) OR (b.visibility='school' AND b.school_id=NULLIF($2,'')::uuid))`, id, c.SchoolID).Scan(&title, &kind, &content, &bookID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 404, "lesson unavailable")
		return
	}
	if err != nil {
		fail(w, 503, "unable to load lesson")
		return
	}
	if kind != "text" && kind != "html" {
		fail(w, 415, "media reader not configured")
		return
	}
	// This returns source as JSON only. The browser displays untrusted HTML as text, not raw markup.
	respond(w, 200, map[string]string{"id": id, "title": title, "content_type": kind, "content": content, "book_id": bookID})
}
func (s *Server) authorBook(w http.ResponseWriter, r *http.Request, bookID string) (auth.Claims, bool) {
	c, exists, err := s.claims(r)
	if err != nil || !exists {
		fail(w, 401, "authentication required")
		return c, false
	}
	if !c.HasAny("platform_admin", "content_admin", "school_admin") {
		fail(w, 403, "editor role required")
		return c, false
	}
	if !validID.MatchString(bookID) {
		fail(w, 400, "invalid book")
		return c, false
	}
	var status, school string
	err = s.db.QueryRowContext(r.Context(), `SELECT status,COALESCE(school_id::text,'') FROM books WHERE id=$1::uuid`, bookID).Scan(&status, &school)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 404, "book not found")
		return c, false
	}
	if err != nil {
		fail(w, 503, "unable to check rights")
		return c, false
	}
	if status != "draft" {
		fail(w, 409, "published content must be revised as a new draft")
		return c, false
	}
	if school == "" {
		if !c.HasRole("platform_admin") {
			fail(w, 403, "platform approval required")
			return c, false
		}
	} else if c.SchoolID != school {
		fail(w, 403, "wrong school")
		return c, false
	}
	return c, true
}
func (s *Server) createChapter(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, ok := s.authorBook(w, r, id)
	if !ok {
		return
	}
	var body struct {
		Title    string `json:"title"`
		Position int    `json:"position"`
	}
	if !decode(w, r, &body) {
		return
	}
	body.Title = strings.TrimSpace(body.Title)
	if len(body.Title) < 2 || len(body.Title) > 180 || body.Position < 1 || body.Position > 500 {
		fail(w, 400, "invalid chapter")
		return
	}
	var created string
	err := s.db.QueryRowContext(r.Context(), `INSERT INTO chapters(book_id,title,position) VALUES($1::uuid,$2,$3) ON CONFLICT(book_id,position) DO NOTHING RETURNING id::text`, id, body.Title, body.Position).Scan(&created)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 409, "position already exists")
		return
	}
	if err != nil {
		fail(w, 503, "unable to add chapter")
		return
	}
	respond(w, 201, map[string]any{"id": created, "title": body.Title, "position": body.Position})
}
func (s *Server) createLesson(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID.MatchString(id) {
		fail(w, 400, "invalid chapter")
		return
	}
	var bookID string
	err := s.db.QueryRowContext(r.Context(), `SELECT book_id::text FROM chapters WHERE id=$1::uuid`, id).Scan(&bookID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 404, "chapter not found")
		return
	}
	if err != nil {
		fail(w, 503, "unable to load chapter")
		return
	}
	_, ok := s.authorBook(w, r, bookID)
	if !ok {
		return
	}
	var body struct {
		Title    string `json:"title"`
		Position int    `json:"position"`
		Content  string `json:"content"`
	}
	if !decode(w, r, &body) {
		return
	}
	body.Title = strings.TrimSpace(body.Title)
	body.Content = strings.TrimSpace(body.Content)
	if len(body.Title) < 2 || len(body.Title) > 180 || body.Position < 1 || body.Position > 500 || len(body.Content) < 2 || len(body.Content) > 50000 {
		fail(w, 400, "invalid lesson")
		return
	}
	var created string
	err = s.db.QueryRowContext(r.Context(), `INSERT INTO lessons(chapter_id,title,position,content_type,content_body) VALUES($1::uuid,$2,$3,'text',$4)
 ON CONFLICT(chapter_id,position) DO NOTHING RETURNING id::text`, id, body.Title, body.Position, body.Content).Scan(&created)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 409, "position already exists")
		return
	}
	if err != nil {
		fail(w, 503, "unable to create lesson")
		return
	}
	respond(w, 201, map[string]any{"id": created, "title": body.Title, "position": body.Position})
}
