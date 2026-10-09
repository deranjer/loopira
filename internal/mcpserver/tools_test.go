package mcpserver

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func uuidFor(t *testing.T, s string) pgtype.UUID {
	t.Helper()
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		t.Fatalf("invalid test uuid %q: %v", s, err)
	}
	return u
}

func TestProjectTargetDate(t *testing.T) {
	got, err := projectTargetDate("2026-09-04")
	if err != nil {
		t.Fatalf("projectTargetDate: %v", err)
	}
	want := time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC)
	if !got.Valid || !got.Time.Equal(want) {
		t.Errorf("got %#v, want valid date %s", got, want)
	}
}

func TestProjectTargetDateAllowsEmptyValue(t *testing.T) {
	got, err := projectTargetDate("")
	if err != nil {
		t.Fatalf("projectTargetDate: %v", err)
	}
	if got.Valid {
		t.Errorf("got %#v, want invalid date", got)
	}
}

func TestProjectTargetDateRejectsInvalidValue(t *testing.T) {
	if _, err := projectTargetDate("September 4"); err == nil {
		t.Fatal("projectTargetDate returned nil error for invalid date")
	}
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

func TestToolsAreTextOnlyWithoutOutputSchema(t *testing.T) {
	ctx := context.Background()
	server := New(nil, nil)
	addTool(server, &mcp.Tool{Name: "echo", Description: "test"},
		func(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, []map[string]int, error) {
			return nil, []map[string]int{{"a": 1}}, nil
		})

	ct, st := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, st, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	session, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer session.Close()

	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	for _, tool := range tools.Tools {
		if tool.OutputSchema != nil {
			t.Errorf("tool %s advertises an output schema; it only costs tokens", tool.Name)
		}
	}

	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "echo"})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.StructuredContent != nil {
		t.Errorf("structuredContent = %v, want none (it duplicates the text)", res.StructuredContent)
	}
	if len(res.Content) != 1 {
		t.Fatalf("got %d content blocks, want 1", len(res.Content))
	}
	if text := res.Content[0].(*mcp.TextContent).Text; text != `[{"a":1}]` {
		t.Errorf("text = %q, want compact JSON", text)
	}
}
