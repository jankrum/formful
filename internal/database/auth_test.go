package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestUpsertUser(t *testing.T) {
	q := testQueries(t)
	ctx := context.Background()

	user, err := q.UpsertUser(ctx, "upsert@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user.Email != "upsert@example.com" {
		t.Fatalf("expected upsert@example.com, got %s", user.Email)
	}

	user2, err := q.UpsertUser(ctx, "upsert@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != user2.ID {
		t.Fatal("upsert returned different ID for same email")
	}
}

func TestGetUserByID(t *testing.T) {
	q := testQueries(t)
	ctx := context.Background()

	created, err := q.UpsertUser(ctx, "getuserbyid@example.com")
	if err != nil {
		t.Fatal(err)
	}

	got, err := q.GetUserByID(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Email != created.Email {
		t.Fatalf("expected %s, got %s", created.Email, got.Email)
	}
}

func TestCreateAndGetMagicLink(t *testing.T) {
	q := testQueries(t)
	ctx := context.Background()

	user, _ := q.UpsertUser(ctx, "magiclink@example.com")

	link, err := q.CreateMagicLink(ctx, CreateMagicLinkParams{
		UserID:    user.ID,
		Token:     "valid-token",
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(15 * time.Minute), Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	row, err := q.GetMagicLinkByToken(ctx, "valid-token")
	if err != nil {
		t.Fatal(err)
	}
	if row.ID != link.ID {
		t.Fatal("returned wrong magic link")
	}
	if row.Email != "magiclink@example.com" {
		t.Fatalf("expected magiclink@example.com, got %s", row.Email)
	}
}

func TestExpiredMagicLinkNotFound(t *testing.T) {
	q := testQueries(t)
	ctx := context.Background()

	user, _ := q.UpsertUser(ctx, "expiredlink@example.com")

	_, err := q.CreateMagicLink(ctx, CreateMagicLinkParams{
		UserID:    user.ID,
		Token:     "expired-token",
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Minute), Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = q.GetMagicLinkByToken(ctx, "expired-token")
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected ErrNoRows for expired token, got %v", err)
	}
}

func TestDeleteMagicLink(t *testing.T) {
	q := testQueries(t)
	ctx := context.Background()

	user, _ := q.UpsertUser(ctx, "deletelink@example.com")

	link, err := q.CreateMagicLink(ctx, CreateMagicLinkParams{
		UserID:    user.ID,
		Token:     "delete-token",
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(15 * time.Minute), Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := q.DeleteMagicLink(ctx, link.ID); err != nil {
		t.Fatal(err)
	}

	_, err = q.GetMagicLinkByToken(ctx, "delete-token")
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected ErrNoRows after delete, got %v", err)
	}
}

func TestCreateAndGetSession(t *testing.T) {
	q := testQueries(t)
	ctx := context.Background()

	user, _ := q.UpsertUser(ctx, "session@example.com")

	session, err := q.CreateSession(ctx, CreateSessionParams{
		ID:        "test-session-id",
		UserID:    user.ID,
		CsrfToken: "test-csrf",
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	row, err := q.GetSessionByID(ctx, session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if row.CsrfToken != "test-csrf" {
		t.Fatalf("expected csrf 'test-csrf', got %s", row.CsrfToken)
	}
	if row.Email != "session@example.com" {
		t.Fatalf("expected session@example.com, got %s", row.Email)
	}
}

func TestExpiredSessionNotFound(t *testing.T) {
	q := testQueries(t)
	ctx := context.Background()

	user, _ := q.UpsertUser(ctx, "expiredsession@example.com")

	_, err := q.CreateSession(ctx, CreateSessionParams{
		ID:        "expired-session-id",
		UserID:    user.ID,
		CsrfToken: "csrf",
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Minute), Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = q.GetSessionByID(ctx, "expired-session-id")
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected ErrNoRows for expired session, got %v", err)
	}
}

func TestDeleteSession(t *testing.T) {
	q := testQueries(t)
	ctx := context.Background()

	user, _ := q.UpsertUser(ctx, "deletesession@example.com")

	_, err := q.CreateSession(ctx, CreateSessionParams{
		ID:        "delete-session-id",
		UserID:    user.ID,
		CsrfToken: "csrf",
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := q.DeleteSession(ctx, "delete-session-id"); err != nil {
		t.Fatal(err)
	}

	_, err = q.GetSessionByID(ctx, "delete-session-id")
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected ErrNoRows after delete, got %v", err)
	}
}
