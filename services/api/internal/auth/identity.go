package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

var ErrCredentials = errors.New("invalid credentials")
var ErrSchoolRequired = errors.New("school selection required")
var ErrLoginLimited = errors.New("too many login attempts")

type LoginResult struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int      `json:"expires_in"`
	UserID      string   `json:"user_id"`
	SchoolID    string   `json:"school_id,omitempty"`
	Roles       []string `json:"roles"`
}

type Identity struct {
	db       *sql.DB
	verifier *Verifier
}

func NewIdentity(db *sql.DB, v *Verifier) *Identity { return &Identity{db: db, verifier: v} }

// Login checks active account status and derives roles from database memberships.
// The caller cannot supply roles. Login failures are throttled in PostgreSQL.
func (i *Identity) Login(ctx context.Context, email, password, schoolID string) (LoginResult, error) {
	var out LoginResult
	email = strings.ToLower(strings.TrimSpace(email))
	if len(email) < 3 || len(email) > 254 || len(password) == 0 || len(password) > 1024 {
		return out, ErrCredentials
	}
	if schoolID != "" && !uuidID.MatchString(schoolID) {
		return out, ErrCredentials
	}
	keyData := sha256.Sum256([]byte(email))
	key := hex.EncodeToString(keyData[:])
	var until sql.NullTime
	err := i.db.QueryRowContext(ctx, `SELECT blocked_until FROM login_failures WHERE key=$1`, key).Scan(&until)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return out, err
	}
	if until.Valid && until.Time.After(time.Now()) {
		return out, ErrLoginLimited
	}
	var userID, hash string
	var authVersion int64
	err = i.db.QueryRowContext(ctx, `SELECT u.id::text,c.password_hash,u.auth_version FROM users u JOIN user_credentials c ON c.user_id=u.id WHERE lower(u.email)=$1 AND u.status='active'`, email).Scan(&userID, &hash, &authVersion)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return out, err
	}
	if err != nil || !VerifyPassword(hash, password) {
		_, e := i.db.ExecContext(ctx, `INSERT INTO login_failures(key,failures,last_attempt,blocked_until) VALUES ($1,1,now(),NULL)
   ON CONFLICT(key) DO UPDATE SET failures=CASE WHEN login_failures.last_attempt < now() - interval '15 minutes' THEN 1 ELSE login_failures.failures+1 END,
   last_attempt=now(), blocked_until=CASE WHEN login_failures.last_attempt >= now() - interval '15 minutes' AND login_failures.failures >= 4 THEN now()+interval '15 minutes' ELSE NULL END`, key)
		if e != nil {
			return out, e
		}
		return out, ErrCredentials
	}
	roles, selected, err := i.rolesFor(ctx, userID, schoolID)
	if err != nil {
		return out, err
	}
	_, _ = i.db.ExecContext(ctx, "DELETE FROM login_failures WHERE key=$1", key)
	claims := Claims{Subject: userID, SchoolID: selected, Roles: roles, AuthVersion: authVersion}
	token, err := i.verifier.Issue(claims, 15*time.Minute)
	if err != nil {
		return out, err
	}
	return LoginResult{AccessToken: token, TokenType: "Bearer", ExpiresIn: 900, UserID: userID, SchoolID: selected, Roles: roles}, nil
}
func (i *Identity) rolesFor(ctx context.Context, userID, schoolID string) ([]string, string, error) {
	if schoolID == "" {
		var platform bool
		if err := i.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM platform_roles WHERE user_id=$1::uuid AND role='platform_admin' AND is_active=true)`, userID).Scan(&platform); err != nil {
			return nil, "", err
		}
		if platform {
			return []string{"platform_admin"}, "", nil
		}
	}
	query := `SELECT school_id::text,role FROM school_memberships WHERE user_id=$1::uuid AND is_active=true`
	args := []any{userID}
	if schoolID != "" {
		query += ` AND school_id=$2::uuid`
		args = append(args, schoolID)
	}
	rows, err := i.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	rolesBySchool := map[string][]string{}
	for rows.Next() {
		var school, role string
		if err := rows.Scan(&school, &role); err != nil {
			return nil, "", err
		}
		rolesBySchool[school] = append(rolesBySchool[school], role)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}
	if len(rolesBySchool) == 0 {
		return nil, "", ErrCredentials
	}
	if len(rolesBySchool) > 1 {
		return nil, "", ErrSchoolRequired
	}
	for s, r := range rolesBySchool {
		return r, s, nil
	}
	return nil, "", ErrCredentials
}

// ValidPrincipal enforces revocation immediately, not only at JWT expiry.
// Every protected read/write checks both active user and active tenant roles.
func (i *Identity) ValidPrincipal(ctx context.Context, c Claims) bool {
	var active bool
	var version int64
	if err := i.db.QueryRowContext(ctx, `SELECT status='active',auth_version FROM users WHERE id=$1::uuid`, c.Subject).Scan(&active, &version); err != nil || !active || c.AuthVersion != version {
		return false
	}
	if c.SchoolID == "" {
		if len(c.Roles) != 1 || c.Roles[0] != "platform_admin" {
			return false
		}
		var ok bool
		err := i.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM platform_roles WHERE user_id=$1::uuid AND role='platform_admin' AND is_active=true)`, c.Subject).Scan(&ok)
		return err == nil && ok
	}
	rows, err := i.db.QueryContext(ctx, `SELECT role FROM school_memberships WHERE school_id=$1::uuid AND user_id=$2::uuid AND is_active=true`, c.SchoolID, c.Subject)
	if err != nil {
		return false
	}
	defer rows.Close()
	allowed := map[string]bool{}
	for rows.Next() {
		var r string
		if rows.Scan(&r) != nil {
			return false
		}
		allowed[r] = true
	}
	if rows.Err() != nil || len(c.Roles) == 0 {
		return false
	}
	seen := map[string]bool{}
	for _, r := range c.Roles {
		if seen[r] || !allowed[r] {
			return false
		}
		seen[r] = true
	}
	return true
}
