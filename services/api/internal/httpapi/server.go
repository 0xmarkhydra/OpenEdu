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

	"github.com/0xmarkhydra/FreeEducation/services/api/internal/auth"
	"github.com/0xmarkhydra/FreeEducation/services/api/internal/catalog"
)

var validID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type Server struct {
	store  catalog.Store
	verify *auth.Verifier
	db     *sql.DB
}

func New(store catalog.Store, verifier *auth.Verifier, db *sql.DB) http.Handler {
	s := &Server{store: store, verify: verifier, db: db}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, _ *http.Request) {
		respond(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /health/ready", s.ready)
	mux.HandleFunc("GET /v1/books", s.list)
	mux.HandleFunc("GET /v1/books/{id}", s.get)
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
