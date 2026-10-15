package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/0xmarkhydra/OpenEdu/services/api/internal/auth"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	SchoolID string `json:"school_id,omitempty"`
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if s.identity == nil {
		fail(w, 503, "identity not configured")
		return
	}
	var input LoginRequest
	if !decode(w, r, &input) {
		return
	}
	result, err := s.identity.Login(r.Context(), strings.TrimSpace(input.Email), input.Password, input.SchoolID)
	if errors.Is(err, auth.ErrSchoolRequired) {
		fail(w, 409, "select a school to sign in")
		return
	}
	if errors.Is(err, auth.ErrLoginLimited) {
		fail(w, 429, "too many attempts")
		return
	}
	if errors.Is(err, auth.ErrCredentials) {
		fail(w, 401, "invalid credentials")
		return
	}
	if err != nil {
		fail(w, 503, "authentication unavailable")
		return
	}
	respond(w, 200, result)
}
func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	claims, has, err := s.claims(r)
	if !has || err != nil {
		fail(w, 401, "authentication required")
		return
	}
	respond(w, 200, map[string]any{"user_id": claims.Subject, "school_id": claims.SchoolID, "roles": claims.Roles})
}
