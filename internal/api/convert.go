package api

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Small conversion helpers from pgtype.* (what sqlc generates) to plain Go
// types (what the JSON API and generated frontend types should look like).
// Kept local to this package for REST-only shapes (api keys, raw uuid
// parsing on input); the shared response DTOs (issues/projects/cycles/
// users) and their own copy of these helpers live in internal/dto so
// internal/mcpserver can reuse them without an import cycle.

const timeFormat = "2006-01-02T15:04:05Z07:00"

func uid(u pgtype.UUID) string {
	return u.String()
}

func nullableUID(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := u.String()
	return &s
}

func nullableText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func ts(t pgtype.Timestamptz) time.Time {
	return t.Time
}

func nullableTimestamp(t pgtype.Timestamptz) *string {
	if !t.Valid {
		return nil
	}
	s := t.Time.Format(timeFormat)
	return &s
}

func mustDate(s string) (pgtype.Date, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Date{}, err
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

func mustUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	err := u.Scan(s)
	return u, err
}
