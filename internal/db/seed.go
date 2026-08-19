package db

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

// EnsureSeedData creates the workspace's first team and admin user on the
// very first boot (when the users table is empty). Self-hosted deployments
// have no signup flow, so this is the only way in.
func EnsureSeedData(ctx context.Context, q *Queries) error {
	count, err := q.CountUsers(ctx)
	if err != nil {
		return fmt.Errorf("counting users: %w", err)
	}
	if count > 0 {
		return nil
	}

	team, err := q.CreateTeam(ctx, CreateTeamParams{Name: "Engineering", Key: "ENG"})
	if err != nil {
		return fmt.Errorf("seeding team: %w", err)
	}

	email := os.Getenv("ADMIN_EMAIL")
	if email == "" {
		email = "admin@loopira.local"
	}
	password := os.Getenv("ADMIN_PASSWORD")
	generated := password == ""
	if generated {
		password = generateRandomPassword()
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing seed password: %w", err)
	}

	user, err := q.CreateUser(ctx, CreateUserParams{
		Name:         "Admin",
		Email:        email,
		PasswordHash: pgtype.Text{String: string(hash), Valid: true},
		Role:         "admin",
	})
	if err != nil {
		return fmt.Errorf("seeding admin user: %w", err)
	}

	if err := q.AddTeamMember(ctx, AddTeamMemberParams{TeamID: team.ID, UserID: user.ID}); err != nil {
		return fmt.Errorf("adding admin to seed team: %w", err)
	}

	// Default labels — same name/color pairs as the pulled design's own
	// LABEL_COLORS — so there's something to assign out of the box.
	defaultLabels := []struct{ name, color string }{
		{"Bug", "#eb5757"},
		{"Chore", "#8a8f98"},
		{"Exploration", "#4fd1c5"},
		{"Polish", "#5e6ad2"},
	}
	for _, l := range defaultLabels {
		if _, err := q.CreateLabel(ctx, CreateLabelParams{TeamID: team.ID, Name: l.name, Color: l.color}); err != nil {
			return fmt.Errorf("seeding label %q: %w", l.name, err)
		}
	}

	slog.Info("seeded workspace", "team", team.Name, "adminEmail", email)
	if generated {
		slog.Warn("generated admin password — save it, it will not be shown again", "password", password)
	}
	return nil
}

func generateRandomPassword() string {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
