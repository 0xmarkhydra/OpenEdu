package httpapi

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/0xmarkhydra/OpenEdu/services/api/internal/auth"
)

// The channel exists only when a teacher is assigned and the guardian link
// to an enrolled child is verified and both accounts have active roles.
const permittedContacts = `SELECT DISTINCT ta.class_id::text,
 CASE WHEN ta.teacher_id=$2::uuid THEN gl.guardian_id::text ELSE ta.teacher_id::text END AS recipient_id,
 CASE WHEN ta.teacher_id=$2::uuid THEN guardian.display_name ELSE teacher.display_name END AS recipient_name,
 cl.name AS class_name
 FROM teacher_assignments ta
 JOIN classes cl ON cl.school_id=ta.school_id AND cl.id=ta.class_id
 JOIN guardian_links gl ON gl.school_id=ta.school_id AND gl.verified_at IS NOT NULL AND gl.revoked_at IS NULL
 JOIN class_students cs ON cs.school_id=ta.school_id AND cs.class_id=ta.class_id AND cs.student_id=gl.student_id
 JOIN school_memberships tm ON tm.school_id=ta.school_id AND tm.user_id=ta.teacher_id AND tm.role='teacher' AND tm.is_active=true
 JOIN school_memberships gm ON gm.school_id=ta.school_id AND gm.user_id=gl.guardian_id AND gm.role='guardian' AND gm.is_active=true
 JOIN users teacher ON teacher.id=ta.teacher_id AND teacher.status='active'
 JOIN users guardian ON guardian.id=gl.guardian_id AND guardian.status='active'
 WHERE ta.school_id=$1::uuid
 AND ((ta.teacher_id=$2::uuid AND $3='teacher') OR (gl.guardian_id=$2::uuid AND $3='guardian'))`

func (s *Server) messageActor(w http.ResponseWriter, r *http.Request) (auth.Claims, string, bool) {
	c, ok, err := s.claims(r)
	if err != nil || !ok {
		fail(w, 401, "authentication required")
		return c, "", false
	}
	if c.SchoolID == "" {
		fail(w, 403, "school membership required")
		return c, "", false
	}
	if c.HasRole("teacher") {
		return c, "teacher", true
	}
	if c.HasRole("guardian") {
		return c, "guardian", true
	}
	fail(w, 403, "teacher or verified guardian required")
	return c, "", false
}
func (s *Server) messageContacts(w http.ResponseWriter, r *http.Request) {
	c, role, ok := s.messageActor(w, r)
	if !ok {
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT class_id,recipient_id,recipient_name,class_name FROM (`+permittedContacts+`) contacts ORDER BY class_name,recipient_name LIMIT 200`, c.SchoolID, c.Subject, role)
	if err != nil {
		fail(w, 503, "unable to load contacts")
		return
	}
	defer rows.Close()
	items := []map[string]string{}
	for rows.Next() {
		var class, id, name, className string
		if err := rows.Scan(&class, &id, &name, &className); err != nil {
			fail(w, 503, "unable to load contacts")
			return
		}
		items = append(items, map[string]string{"class_id": class, "recipient_id": id, "name": name, "class_name": className})
	}
	if rows.Err() != nil {
		fail(w, 503, "unable to load contacts")
		return
	}
	respond(w, 200, map[string]any{"items": items})
}
func (s *Server) messages(w http.ResponseWriter, r *http.Request) {
	c, role, ok := s.messageActor(w, r)
	if !ok {
		return
	}
	if r.Method == "POST" {
		var input struct {
			ClassID     string `json:"class_id"`
			RecipientID string `json:"recipient_id"`
			Content     string `json:"content"`
		}
		if !decode(w, r, &input) {
			return
		}
		input.Content = strings.TrimSpace(input.Content)
		if !validID.MatchString(input.ClassID) || !validID.MatchString(input.RecipientID) || len(input.Content) == 0 || len(input.Content) > 2000 || input.RecipientID == c.Subject {
			fail(w, 400, "invalid message")
			return
		}
		// Enforce recipient eligibility with the same verified-contact query as GET.
		var eligible bool
		err := s.db.QueryRowContext(r.Context(), `SELECT EXISTS(
   SELECT 1 FROM (`+permittedContacts+`) allowed
   WHERE allowed.class_id=$4::text AND allowed.recipient_id=$5::text)`, c.SchoolID, c.Subject, role, input.ClassID, input.RecipientID).Scan(&eligible)
		if err != nil {
			fail(w, 503, "unable to check contact")
			return
		}
		if !eligible {
			fail(w, 403, "recipient is not an approved contact")
			return
		}
		var id string
		// Repeat approval check at insertion time to block forged IDs or role revocation.
		err = s.db.QueryRowContext(r.Context(), `INSERT INTO school_messages(school_id,class_id,sender_id,recipient_id,content)
   SELECT $1::uuid,$4::uuid,$2::uuid,$5::uuid,$6
   WHERE EXISTS(SELECT 1 FROM (`+permittedContacts+`) allowed
     WHERE allowed.class_id=$4::text AND allowed.recipient_id=$5::text)
   RETURNING id::text`, c.SchoolID, c.Subject, role, input.ClassID, input.RecipientID, input.Content).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			fail(w, 403, "contact no longer available")
			return
		}
		if err != nil {
			fail(w, 503, "unable to send message")
			return
		}
		respond(w, 201, map[string]string{"id": id, "status": "sent"})
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT m.id::text,m.class_id::text,m.sender_id::text,m.recipient_id::text,m.content,m.created_at,m.read_at
 FROM school_messages m WHERE m.school_id=$1::uuid AND (m.sender_id=$2::uuid OR m.recipient_id=$2::uuid)
 AND EXISTS(SELECT 1 FROM (`+permittedContacts+`) allowed
 WHERE allowed.class_id=m.class_id::text AND
 allowed.recipient_id = CASE WHEN m.sender_id=$2::uuid THEN m.recipient_id::text ELSE m.sender_id::text END)
 ORDER BY m.created_at DESC LIMIT 100`, c.SchoolID, c.Subject, role)
	if err != nil {
		fail(w, 503, "unable to load messages")
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, class, sender, recipient, content string
		var sent time.Time
		var read sql.NullTime
		if err := rows.Scan(&id, &class, &sender, &recipient, &content, &sent, &read); err != nil {
			fail(w, 503, "unable to load messages")
			return
		}
		var at any
		if read.Valid {
			at = read.Time
		}
		items = append(items, map[string]any{"id": id, "class_id": class, "sender_id": sender, "recipient_id": recipient, "content": content, "created_at": sent, "read_at": at})
	}
	if rows.Err() != nil {
		fail(w, 503, "unable to load messages")
		return
	}
	respond(w, 200, map[string]any{"items": items})
}
func (s *Server) readMessage(w http.ResponseWriter, r *http.Request) {
	c, role, ok := s.messageActor(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if !validID.MatchString(id) {
		fail(w, 400, "invalid message id")
		return
	}
	result, err := s.db.ExecContext(r.Context(), `UPDATE school_messages m SET read_at=COALESCE(m.read_at,now())
 WHERE m.id=$4::uuid AND m.school_id=$1::uuid AND m.recipient_id=$2::uuid
 AND EXISTS(SELECT 1 FROM (`+permittedContacts+`) allowed
 WHERE allowed.class_id=m.class_id::text AND allowed.recipient_id=m.sender_id::text)`, c.SchoolID, c.Subject, role, id)
	if err != nil {
		fail(w, 503, "unable to update receipt")
		return
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		fail(w, 404, "message unavailable")
		return
	}
	respond(w, 200, map[string]string{"status": "read"})
}
