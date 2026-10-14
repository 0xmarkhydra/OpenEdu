package catalog

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrNotFound  = errors.New("resource not found")
	ErrInvalid   = errors.New("invalid input")
	ErrForbidden = errors.New("forbidden")
	ErrConflict  = errors.New("invalid status transition")
)

type Actor struct {
	UserID   string
	SchoolID string
	Global   bool
}

type BookInput struct {
	Title         string `json:"title"`
	Grade         int    `json:"grade"`
	SubjectCode   string `json:"subject_code"`
	RightsOwner   string `json:"rights_owner"`
	RightsLicense string `json:"rights_license"`
	RightsSource  string `json:"rights_source"`
	Visibility    string `json:"visibility"`
}

type Book struct {
	ID       string `json:"id"`
	SchoolID string `json:"school_id,omitempty"`
	BookInput
	Status      string     `json:"status"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

type Filter struct {
	SchoolID string
	Grade    int
	Subject  string
	Query    string
	Limit    int
}

type Store interface {
	List(context.Context, Filter) ([]Book, error)
	Get(context.Context, string, string) (Book, error)
	Create(context.Context, Actor, BookInput) (Book, error)
	Transition(context.Context, Actor, string, string, string) error
}

func (input BookInput) Validate(actor Actor) error {
	for _, s := range []string{input.Title, input.SubjectCode, input.RightsOwner, input.RightsLicense, input.RightsSource} {
		if len(strings.TrimSpace(s)) == 0 || len(s) > 500 {
			return ErrInvalid
		}
	}
	if input.Grade < 1 || input.Grade > 12 {
		return ErrInvalid
	}
	switch input.Visibility {
	case "public":
		if !actor.Global {
			return ErrForbidden
		}
	case "school":
		if actor.SchoolID == "" {
			return ErrForbidden
		}
	default:
		return ErrInvalid
	}
	if actor.UserID == "" {
		return ErrForbidden
	}
	return nil
}

func NextStatus(action string) (from, to string, err error) {
	switch action {
	case "submit":
		return "draft", "review", nil
	case "approve":
		return "review", "approved", nil
	case "publish":
		return "approved", "published", nil
	case "archive":
		return "published", "archived", nil
	default:
		return "", "", ErrInvalid
	}
}
