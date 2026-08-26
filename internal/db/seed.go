package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

// SetupRequired reports whether the workspace has never been set up — no
// users exist yet, so the /setup wizard should run. Self-hosted deployments
// have no signup flow, so this is the only gate on first-run state; once an
// admin user exists it can never report true again.
func SetupRequired(ctx context.Context, q *Queries) (bool, error) {
	count, err := q.CountUsers(ctx)
	if err != nil {
		return false, fmt.Errorf("counting users: %w", err)
	}
	return count == 0, nil
}

// CompleteSetup creates the workspace's first team and admin user. Callers
// must check SetupRequired first — this does not re-check, so it must only
// run from the one-time /setup flow while the users table is still empty.
func CompleteSetup(ctx context.Context, q *Queries, name, email, password string) (User, error) {
	team, err := q.CreateTeam(ctx, CreateTeamParams{Name: "Engineering", Key: "ENG"})
	if err != nil {
		return User{}, fmt.Errorf("creating team: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("hashing password: %w", err)
	}

	user, err := q.CreateUser(ctx, CreateUserParams{
		Name:         name,
		Email:        email,
		PasswordHash: pgtype.Text{String: string(hash), Valid: true},
		Role:         "admin",
	})
	if err != nil {
		return User{}, fmt.Errorf("creating admin user: %w", err)
	}

	if err := q.AddTeamMember(ctx, AddTeamMemberParams{TeamID: team.ID, UserID: user.ID}); err != nil {
		return User{}, fmt.Errorf("adding admin to team: %w", err)
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
			return User{}, fmt.Errorf("seeding label %q: %w", l.name, err)
		}
	}

	return user, nil
}
