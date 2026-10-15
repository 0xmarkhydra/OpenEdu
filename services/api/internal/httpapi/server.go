package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/0xmarkhydra/OpenEdu/services/api/internal/auth"
	"github.com/0xmarkhydra/OpenEdu/services/api/internal/catalog"
)

var validID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type Server struct {
	store    catalog.Store
	verify   *auth.Verifier
	db       *sql.DB
	identity *auth.Identity
}

func New(store catalog.Store, verifier *auth.Verifier, db *sql.DB) http.Handler {
	s := &Server{store: store, verify: verifier, db: db}
	if db != nil {
		s.identity = auth.NewIdentity(db, verifier)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, _ *http.Request) {
		respond(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /health/ready", s.ready)
	mux.HandleFunc("POST /v1/auth/login", s.login)
	mux.HandleFunc("GET /v1/auth/me", s.me)
	mux.HandleFunc("POST /v1/auth/password", s.changePassword)
	mux.HandleFunc("GET /v1/admin/schools", s.schools)
	mux.HandleFunc("POST /v1/admin/schools", s.schools)
	mux.HandleFunc("POST /v1/admin/schools/{id}/members", s.addSchoolMember)
	mux.HandleFunc("GET /v1/schools/{school}/years", s.years)
	mux.HandleFunc("POST /v1/schools/{school}/years", s.years)
	mux.HandleFunc("GET /v1/schools/{school}/classes", s.classes)
	mux.HandleFunc("POST /v1/schools/{school}/classes", s.classes)
	mux.HandleFunc("GET /v1/schools/{school}/announcements", s.announcements)
	mux.HandleFunc("POST /v1/schools/{school}/announcements", s.announcements)
	mux.HandleFunc("GET /v1/schools/{school}/overview", s.overview)
	mux.HandleFunc("GET /v1/schools/{school}/reports", s.schoolReport)
	mux.HandleFunc("GET /v1/classes/{id}/students", s.students)
	mux.HandleFunc("POST /v1/classes/{id}/students", s.students)
	mux.HandleFunc("GET /v1/classes/{id}/attendance", s.attendance)
	mux.HandleFunc("PUT /v1/classes/{id}/attendance", s.attendance)
	mux.HandleFunc("POST /v1/classes/{id}/teachers", s.assignTeacher)
	mux.HandleFunc("GET /v1/classes/{id}/timetable", s.timetable)
	mux.HandleFunc("POST /v1/classes/{id}/timetable", s.timetable)
	mux.HandleFunc("GET /v1/classes/{id}/quizzes", s.classQuizzes)
	mux.HandleFunc("POST /v1/classes/{id}/quizzes", s.classQuizzes)
	mux.HandleFunc("GET /v1/quizzes/{id}", s.quizDetail)
	mux.HandleFunc("POST /v1/quizzes/{id}/questions", s.quizQuestions)
	mux.HandleFunc("POST /v1/quizzes/{id}/publish", s.publishQuiz)
	mux.HandleFunc("POST /v1/quizzes/{id}/start", s.startQuiz)
	mux.HandleFunc("POST /v1/quizzes/{id}/submit", s.submitQuiz)
	mux.HandleFunc("GET /v1/quizzes/{id}/results", s.quizResults)
	mux.HandleFunc("GET /v1/classes/{id}/assignments", s.assignments)
	mux.HandleFunc("POST /v1/classes/{id}/assignments", s.assignments)
	mux.HandleFunc("GET /v1/assignments/{id}/submissions", s.submissions)
	mux.HandleFunc("POST /v1/assignments/{id}/submissions", s.submissions)
	mux.HandleFunc("PATCH /v1/assignments/{id}/submissions", s.submissions)
	mux.HandleFunc("GET /v1/classes/{id}/leave-requests", s.leaveRequests)
	mux.HandleFunc("POST /v1/classes/{id}/leave-requests", s.leaveRequests)
	mux.HandleFunc("PATCH /v1/classes/{id}/leave-requests", s.leaveRequests)
	mux.HandleFunc("POST /v1/schools/{school}/guardians", s.grantGuardian)
	mux.HandleFunc("GET /v1/me/messages/contacts", s.messageContacts)
	mux.HandleFunc("GET /v1/me/messages", s.messages)
	mux.HandleFunc("POST /v1/me/messages", s.messages)
	mux.HandleFunc("PATCH /v1/me/messages/{id}/read", s.readMessage)
	mux.HandleFunc("GET /v1/me/children", s.children)
	mux.HandleFunc("GET /v1/me/children/{student}/attendance", s.childAttendance)
	mux.HandleFunc("GET /v1/me/children/{student}/grades", s.childGrades)
	mux.HandleFunc("POST /v1/me/children/{student}/leave", s.childLeave)
	mux.HandleFunc("GET /v1/books/{id}/outline", s.outline)
	mux.HandleFunc("GET /v1/books/{id}/progress", s.bookProgress)
	mux.HandleFunc("PUT /v1/lessons/{id}/progress", s.lessonProgress)
	mux.HandleFunc("GET /v1/lessons/{id}", s.lessonContent)
	mux.HandleFunc("POST /v1/admin/books/{id}/chapters", s.createChapter)
	mux.HandleFunc("POST /v1/admin/chapters/{id}/lessons", s.createLesson)
	mux.HandleFunc("GET /v1/books", s.list)
	mux.HandleFunc("GET /v1/books/{id}", s.get)
	mux.HandleFunc("GET /v1/admin/books", s.adminBooks)
	mux.HandleFunc("GET /v1/admin/books/{id}/chapters", s.draftOutline)
	mux.HandleFunc("POST /v1/admin/books", s.create)
	mux.HandleFunc("POST /v1/admin/books/{id}/{action}", s.transition)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		mux.ServeHTTP(w, r)
	})
}
func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, status int, detail string) {
	respond(w, status, map[string]string{"error": detail})
}
func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	if s.db == nil {
		fail(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.db.PingContext(ctx); err != nil {
		fail(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "ready"})
}
func (s *Server) claims(r *http.Request) (auth.Claims, bool, error) {
	value := r.Header.Get("Authorization")
	if value == "" {
		return auth.Claims{}, false, nil
	}
	if !strings.HasPrefix(value, "Bearer ") {
		return auth.Claims{}, false, auth.ErrInvalidToken
	}
	c, err := s.verify.Verify(strings.TrimSpace(strings.TrimPrefix(value, "Bearer ")))
	if err != nil {
		return auth.Claims{}, false, err
	}
	if s.identity != nil && !s.identity.ValidPrincipal(r.Context(), c) {
		return auth.Claims{}, false, auth.ErrInvalidToken
	}
	return c, true, nil
}
func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	c, _, err := s.claims(r)
	if err != nil {
		fail(w, http.StatusUnauthorized, "invalid token")
		return
	}
	grade := 0
	if raw := r.URL.Query().Get("grade"); raw != "" {
		grade, err = strconv.Atoi(raw)
		if err != nil || grade < 1 || grade > 12 {
			fail(w, 400, "invalid grade")
			return
		}
	}
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 100 {
			fail(w, 400, "invalid limit")
			return
		}
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	subject := strings.TrimSpace(r.URL.Query().Get("subject"))
	if len(q) > 120 || len(subject) > 80 {
		fail(w, 400, "query too long")
		return
	}
	list, err := s.store.List(r.Context(), catalog.Filter{SchoolID: c.SchoolID, Grade: grade, Subject: subject, Query: q, Limit: limit})
	if err != nil {
		fail(w, 500, "unable to retrieve books")
		return
	}
	respond(w, 200, map[string]any{"items": list})
}
func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !validID.MatchString(id) {
		fail(w, 400, "invalid id")
		return
	}
	c, _, err := s.claims(r)
	if err != nil {
		fail(w, 401, "invalid token")
		return
	}
	b, err := s.store.Get(r.Context(), c.SchoolID, id)
	if errors.Is(err, catalog.ErrNotFound) {
		fail(w, 404, "book not found")
		return
	}
	if err != nil {
		fail(w, 500, "unable to retrieve book")
		return
	}
	respond(w, 200, b)
}
func (s *Server) authorized(r *http.Request) (auth.Claims, bool) {
	c, exists, err := s.claims(r)
	return c, exists && err == nil
}
func actorFor(c auth.Claims) catalog.Actor {
	return catalog.Actor{UserID: c.Subject, SchoolID: c.SchoolID, Global: c.SchoolID == "" && c.HasAny("platform_admin", "content_admin", "content_reviewer")}
}
func decode(w http.ResponseWriter, r *http.Request, out any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 65536)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		fail(w, 400, "invalid json request")
		return false
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		fail(w, 400, "unexpected trailing body")
		return false
	}
	return true
}
func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	c, ok := s.authorized(r)
	if !ok {
		fail(w, 401, "authentication required")
		return
	}
	if !c.HasAny("platform_admin", "content_admin", "school_admin") {
		fail(w, 403, "insufficient permissions")
		return
	}
	var body catalog.BookInput
	if !decode(w, r, &body) {
		return
	}
	book, err := s.store.Create(r.Context(), actorFor(c), body)
	if errors.Is(err, catalog.ErrForbidden) {
		fail(w, 403, "wrong tenant or visibility")
		return
	}
	if errors.Is(err, catalog.ErrInvalid) {
		fail(w, 400, "invalid book metadata")
		return
	}
	if err != nil {
		fail(w, 500, "unable to create book")
		return
	}
	respond(w, http.StatusCreated, book)
}
func (s *Server) transition(w http.ResponseWriter, r *http.Request) {
	c, ok := s.authorized(r)
	if !ok {
		fail(w, 401, "authentication required")
		return
	}
	id := r.PathValue("id")
	if !validID.MatchString(id) {
		fail(w, 400, "invalid id")
		return
	}
	action := r.PathValue("action")
	from, to, err := catalog.NextStatus(action)
	if err != nil {
		fail(w, 404, "unknown action")
		return
	}
	switch action {
	case "submit":
		ok = c.HasAny("school_admin", "content_admin", "platform_admin")
	case "approve":
		ok = c.HasAny("content_reviewer", "platform_admin")
	case "publish", "archive":
		ok = c.HasAny("school_admin", "content_admin", "platform_admin")
	}
	if !ok {
		fail(w, 403, "insufficient permissions")
		return
	}
	err = s.store.Transition(r.Context(), actorFor(c), id, from, to)
	switch {
	case errors.Is(err, catalog.ErrNotFound):
		fail(w, 404, "book not found")
	case errors.Is(err, catalog.ErrConflict):
		fail(w, 409, "invalid publication transition")
	case errors.Is(err, catalog.ErrForbidden):
		fail(w, 403, "insufficient permissions")
	case err != nil:
		fail(w, 500, "unable to update book")
	default:
		respond(w, 200, map[string]string{"id": id, "status": to})
	}
}
