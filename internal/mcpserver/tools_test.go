package mcpserver

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func uuidFor(t *testing.T, s string) pgtype.UUID {
	t.Helper()
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		t.Fatalf("invalid test uuid %q: %v", s, err)
	}
	return u
}

func TestMergeOptionalRefNilLeavesUnchanged(t *testing.T) {
	current := uuidFor(t, "11111111-1111-1111-1111-111111111111")
	got, err := mergeOptionalRef(current, nil)
	if err != nil {
		t.Fatalf("mergeOptionalRef: %v", err)
	}
	if got != current {
		t.Errorf("got %v, want unchanged %v", got, current)
	}
}

func TestMergeOptionalRefEmptyStringClears(t *testing.T) {
	current := uuidFor(t, "11111111-1111-1111-1111-111111111111")
	empty := ""
	got, err := mergeOptionalRef(current, &empty)
	if err != nil {
		t.Fatalf("mergeOptionalRef: %v", err)
	}
	if got.Valid {
		t.Errorf("got %v, want a cleared (invalid) uuid", got)
	}
}

func TestMergeOptionalRefValidUUIDIsParsed(t *testing.T) {
	current := uuidFor(t, "11111111-1111-1111-1111-111111111111")
	next := "22222222-2222-2222-2222-222222222222"
	got, err := mergeOptionalRef(current, &next)
	if err != nil {
		t.Fatalf("mergeOptionalRef: %v", err)
	}
	want := uuidFor(t, next)
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMergeOptionalRefInvalidUUIDErrors(t *testing.T) {
	current := uuidFor(t, "11111111-1111-1111-1111-111111111111")
	next := "not-a-uuid"
	if _, err := mergeOptionalRef(current, &next); err == nil {
		t.Error("mergeOptionalRef with an invalid uuid string returned nil error, want an error")
	}
}
