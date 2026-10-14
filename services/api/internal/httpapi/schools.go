package httpapi

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/0xmarkhydra/FreeEducation/services/api/internal/auth"
)

var schoolSlug = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,48}[a-z0-9]$`)

type School struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (s *Server) platformAdmin(w http.ResponseWriter, r *http.Request) bool {
	c, exists, err := s.claims(r)
	if !exists || err != nil {
		fail(w, 401, "authentication required")
		return false
	}
	if c.SchoolID != "" || !c.HasRole("platform_admin") {
		fail(w, 403, "platform administrator required")
		return false
	}
	return true
}

func (s *Server) schools(w http.ResponseWriter, r *http.Request) {
	if s.db == nil {
		fail(w, 503, "database unavailable")
		return
	}
	if !s.platformAdmin(w, r) {
		return
	}
	if r.Method == "GET" {
		rows, err := s.db.QueryContext(r.Context(), `SELECT id::text,name,slug FROM schools ORDER BY name LIMIT 500`)
		if err != nil {
			fail(w, 503, "unable to read schools")
			return
		}
		defer rows.Close()
		schools := []School{}
		for rows.Next() {
			var school School
			if err := rows.Scan(&school.ID, &school.Name, &school.Slug); err != nil {
				fail(w, 503, "unable to read schools")
				return
			}
			schools = append(schools, school)
		}
		if rows.Err() != nil {
			fail(w, 503, "unable to read schools")
			return
		}
		respond(w, 200, map[string]any{"items": schools})
		return
	}
	var input struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if !decode(w, r, &input) {
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.TrimSpace(strings.ToLower(input.Slug))
	if len(input.Name) < 2 || len(input.Name) > 160 || !schoolSlug.MatchString(input.Slug) {
		fail(w, 400, "invalid school name or slug")
		return
	}
	var school School
	err := s.db.QueryRowContext(r.Context(), `INSERT INTO schools(name,slug) VALUES($1,$2) ON CONFLICT(slug) DO NOTHING RETURNING id::text,name,slug`, input.Name, input.Slug).Scan(&school.ID, &school.Name, &school.Slug)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 409, "school slug already exists")
		return
	}
	if err != nil {
		fail(w, 503, "unable to create school")
		return
	}
	respond(w, 201, school)
}

func (s *Server) addSchoolMember(w http.ResponseWriter, r *http.Request) {
	if s.db == nil {
		fail(w, 503, "database unavailable")
		return
	}
	if !s.platformAdmin(w, r) {
		return
	}
	schoolID := r.PathValue("id")
	if !validID.MatchString(schoolID) {
		fail(w, 400, "invalid school")
		return
	}
	var input struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if !decode(w, r, &input) {
		return
	}
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.Name = strings.TrimSpace(input.Name)
	if !strings.Contains(input.Email, "@") || len(input.Email) > 254 || len(input.Name) < 2 || len(input.Name) > 160 {
		fail(w, 400, "invalid user")
		return
	}
	if input.Role != "school_admin" && input.Role != "teacher" && input.Role != "student" && input.Role != "guardian" && input.Role != "principal" && input.Role != "content_admin" && input.Role != "content_reviewer" {
		fail(w, 400, "invalid role")
		return
	}
	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		fail(w, 400, "password must be 12-72 bytes")
		return
	}
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		fail(w, 503, "unable to create account")
		return
	}
	defer tx.Rollback()
	var schoolExists bool
	err = tx.QueryRowContext(r.Context(), `SELECT EXISTS(SELECT 1 FROM schools WHERE id=$1::uuid)`, schoolID).Scan(&schoolExists)
	if err != nil {
		fail(w, 503, "unable to create account")
		return
	}
	if !schoolExists {
		fail(w, 404, "school not found")
		return
	}
	var userID string
	err = tx.QueryRowContext(r.Context(), `INSERT INTO users(email,display_name) VALUES($1,$2) ON CONFLICT DO NOTHING RETURNING id::text`, input.Email, input.Name).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, 409, "account already exists")
		return
	}
	if err != nil {
		fail(w, 503, "unable to create account")
		return
	}
	_, err = tx.ExecContext(r.Context(), `INSERT INTO user_credentials(user_id,password_hash) VALUES($1::uuid,$2)`, userID, hash)
	if err != nil {
		fail(w, 503, "unable to create account")
		return
	}
	_, err = tx.ExecContext(r.Context(), `INSERT INTO school_memberships(school_id,user_id,role) VALUES($1::uuid,$2::uuid,$3)`, schoolID, userID, input.Role)
	if err != nil {
		fail(w, 503, "unable to create account")
		return
	}
	if err = tx.Commit(); err != nil {
		fail(w, 503, "unable to create account")
		return
	}
	respond(w, 201, map[string]string{"user_id": userID, "school_id": schoolID, "role": input.Role})
}
