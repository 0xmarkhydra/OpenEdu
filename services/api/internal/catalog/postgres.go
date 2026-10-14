package catalog

import (
	"context"
	"database/sql"
	"errors"
)

type PostgresStore struct{ db *sql.DB }

func NewPostgresStore(db *sql.DB) *PostgresStore { return &PostgresStore{db: db} }

const selectBook = "SELECT id::text,COALESCE(school_id::text,''),title,grade,subject_code,rights_owner,rights_license,rights_source,visibility,status,published_at FROM books"

func readBook(scanner interface{ Scan(...any) error }) (Book, error) {
	var b Book
	var when sql.NullTime
	err := scanner.Scan(&b.ID, &b.SchoolID, &b.Title, &b.Grade, &b.SubjectCode, &b.RightsOwner, &b.RightsLicense, &b.RightsSource, &b.Visibility, &b.Status, &when)
	if when.Valid {
		b.PublishedAt = &when.Time
	}
	return b, err
}

func (s *PostgresStore) List(ctx context.Context, f Filter) ([]Book, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	rows, err := s.db.QueryContext(ctx, selectBook+` WHERE status='published'
	AND ((visibility='public' AND school_id IS NULL) OR (visibility='school' AND school_id=NULLIF($1,'')::uuid))
	AND ($2='' OR title ILIKE ('%'||$2||'%'))
	AND ($3=0 OR grade=$3)
	AND ($4='' OR subject_code=$4)
	ORDER BY title,id LIMIT $5`, f.SchoolID, f.Query, f.Grade, f.Subject, f.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	books := make([]Book, 0)
	for rows.Next() {
		b, err := readBook(rows)
		if err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, rows.Err()
}
func (s *PostgresStore) Get(ctx context.Context, schoolID, id string) (Book, error) {
	b, err := readBook(s.db.QueryRowContext(ctx, selectBook+` WHERE id=$1::uuid AND status='published'
	AND ((visibility='public' AND school_id IS NULL) OR (visibility='school' AND school_id=NULLIF($2,'')::uuid))`, id, schoolID))
	if errors.Is(err, sql.ErrNoRows) {
		return Book{}, ErrNotFound
	}
	return b, err
}
func (s *PostgresStore) Create(ctx context.Context, actor Actor, input BookInput) (Book, error) {
	if err := input.Validate(actor); err != nil {
		return Book{}, err
	}
	var id string
	err := s.db.QueryRowContext(ctx, `INSERT INTO books
	(school_id,title,grade,subject_code,rights_owner,rights_license,rights_source,visibility,created_by)
	VALUES(NULLIF($1,'')::uuid,$2,$3,$4,$5,$6,$7,$8,$9::uuid) RETURNING id::text`,
		func() string {
			if input.Visibility == "public" {
				return ""
			}
			return actor.SchoolID
		}(), input.Title, input.Grade, input.SubjectCode, input.RightsOwner, input.RightsLicense, input.RightsSource, input.Visibility, actor.UserID).Scan(&id)
	if err != nil {
		return Book{}, err
	}
	school := ""
	if input.Visibility == "school" {
		school = actor.SchoolID
	}
	return Book{ID: id, SchoolID: school, BookInput: input, Status: "draft"}, nil
}
func (s *PostgresStore) Transition(ctx context.Context, actor Actor, id, from, to string) error {
	if actor.UserID == "" {
		return ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var current, school string
	err = tx.QueryRowContext(ctx, `SELECT status,COALESCE(school_id::text,'') FROM books WHERE id=$1::uuid FOR UPDATE`, id).Scan(&current, &school)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if !actor.Global && (school == "" || school != actor.SchoolID) {
		return ErrNotFound
	}
	if current != from {
		return ErrConflict
	}
	_, err = tx.ExecContext(ctx, `UPDATE books SET status=$2,updated_at=now(),published_at=CASE WHEN $2='published' THEN now() ELSE published_at END WHERE id=$1::uuid`, id, to)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO book_audit_log(book_id,actor_id,from_status,to_status) VALUES($1::uuid,$2::uuid,$3,$4)`, id, actor.UserID, from, to)
	if err != nil {
		return err
	}
	return tx.Commit()
}
