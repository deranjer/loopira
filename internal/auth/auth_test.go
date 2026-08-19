package auth

import (
	"context"
	"testing"
)

func TestIdentityContextRoundTrip(t *testing.T) {
	ctx := context.Background()

	if _, ok := IdentityFromContext(ctx); ok {
		t.Error("IdentityFromContext returned ok=true for an empty context")
	}
	if _, ok := UserID(ctx); ok {
		t.Error("UserID returned ok=true for an empty context")
	}
	if CanWrite(ctx) {
		t.Error("CanWrite returned true for an empty context")
	}
	if ViaAPIKey(ctx) {
		t.Error("ViaAPIKey returned true for an empty context")
	}

	sessionCtx := WithIdentity(ctx, Identity{UserID: "user-1", Write: true})
	id, ok := IdentityFromContext(sessionCtx)
	if !ok || id.UserID != "user-1" || !id.Write || id.ViaAPIKey {
		t.Errorf("session identity roundtrip = %+v, ok=%v", id, ok)
	}
	if userID, ok := UserID(sessionCtx); !ok || userID != "user-1" {
		t.Errorf("UserID(sessionCtx) = %q, %v", userID, ok)
	}
	if !CanWrite(sessionCtx) {
		t.Error("CanWrite(sessionCtx) = false, want true (sessions are always full access)")
	}
	if ViaAPIKey(sessionCtx) {
		t.Error("ViaAPIKey(sessionCtx) = true, want false")
	}

	readOnlyKeyCtx := WithIdentity(ctx, Identity{UserID: "user-2", ViaAPIKey: true, Write: false})
	if CanWrite(readOnlyKeyCtx) {
		t.Error("CanWrite(readOnlyKeyCtx) = true, want false")
	}
	if !ViaAPIKey(readOnlyKeyCtx) {
		t.Error("ViaAPIKey(readOnlyKeyCtx) = false, want true")
	}

	writeKeyCtx := WithIdentity(ctx, Identity{UserID: "user-3", ViaAPIKey: true, Write: true})
	if !CanWrite(writeKeyCtx) {
		t.Error("CanWrite(writeKeyCtx) = false, want true")
	}
}
