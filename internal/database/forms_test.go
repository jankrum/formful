package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func testUser(t *testing.T, q *Queries, email string) User {
	t.Helper()
	user, err := q.UpsertUser(context.Background(), email)
	if err != nil {
		t.Fatal(err)
	}
	return user
}

func TestCreateAndGetForm(t *testing.T) {
	q := testQueries(t)
	ctx := context.Background()

	user := testUser(t, q, "formcreate@example.com")

	form, err := q.CreateForm(ctx, CreateFormParams{
		UserID:      user.ID,
		Title:       "My Form",
		Description: "A test form",
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := q.GetFormByID(ctx, form.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "My Form" || got.Description != "A test form" {
		t.Fatalf("unexpected form: %+v", got)
	}
}

func TestListFormsByUserID(t *testing.T) {
	q := testQueries(t)
	ctx := context.Background()

	user := testUser(t, q, "formlist@example.com")

	for _, title := range []string{"Form A", "Form B", "Form C"} {
		_, err := q.CreateForm(ctx, CreateFormParams{UserID: user.ID, Title: title})
		if err != nil {
			t.Fatal(err)
		}
	}

	forms, err := q.ListFormsByUserID(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(forms) != 3 {
		t.Fatalf("expected 3 forms, got %d", len(forms))
	}
}

func TestUpdateForm(t *testing.T) {
	q := testQueries(t)
	ctx := context.Background()

	user := testUser(t, q, "formupdate@example.com")

	form, err := q.CreateForm(ctx, CreateFormParams{UserID: user.ID, Title: "Old Title"})
	if err != nil {
		t.Fatal(err)
	}

	updated, err := q.UpdateForm(ctx, UpdateFormParams{
		ID:          form.ID,
		Title:       "New Title",
		Description: "New description",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "New Title" {
		t.Fatalf("expected 'New Title', got %s", updated.Title)
	}
}

func TestDeleteForm(t *testing.T) {
	q := testQueries(t)
	ctx := context.Background()

	user := testUser(t, q, "formdelete@example.com")

	form, err := q.CreateForm(ctx, CreateFormParams{UserID: user.ID, Title: "To Delete"})
	if err != nil {
		t.Fatal(err)
	}

	if err := q.DeleteForm(ctx, form.ID); err != nil {
		t.Fatal(err)
	}

	_, err = q.GetFormByID(ctx, form.ID)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected ErrNoRows after delete, got %v", err)
	}
}

func TestCreateAndListQuestions(t *testing.T) {
	q := testQueries(t)
	ctx := context.Background()

	user := testUser(t, q, "questions@example.com")
	form, _ := q.CreateForm(ctx, CreateFormParams{UserID: user.ID, Title: "Question Form"})

	questionDefs := []struct {
		title    string
		qtype    QuestionType
		position int32
	}{
		{"Name", QuestionTypeShortText, 1},
		{"Bio", QuestionTypeLongText, 2},
		{"Favorite color", QuestionTypeMultipleChoice, 3},
	}

	for _, d := range questionDefs {
		_, err := q.CreateQuestion(ctx, CreateQuestionParams{
			FormID:   form.ID,
			Type:     d.qtype,
			Title:    d.title,
			Position: d.position,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	questions, err := q.ListQuestionsByFormID(ctx, form.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(questions) != 3 {
		t.Fatalf("expected 3 questions, got %d", len(questions))
	}
	if questions[0].Title != "Name" || questions[1].Title != "Bio" {
		t.Fatal("questions not in position order")
	}
}

func TestResponseAndAnswers(t *testing.T) {
	q := testQueries(t)
	ctx := context.Background()

	user := testUser(t, q, "responses@example.com")
	form, _ := q.CreateForm(ctx, CreateFormParams{UserID: user.ID, Title: "Response Form"})
	question, _ := q.CreateQuestion(ctx, CreateQuestionParams{
		FormID:   form.ID,
		Type:     QuestionTypeShortText,
		Title:    "What is your name?",
		Position: 1,
	})

	response, err := q.CreateResponse(ctx, form.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = q.CreateAnswer(ctx, CreateAnswerParams{
		ResponseID: response.ID,
		QuestionID: question.ID,
		Value:      "Alice",
	})
	if err != nil {
		t.Fatal(err)
	}

	answers, err := q.ListAnswersByResponseID(ctx, response.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(answers) != 1 || answers[0].Value != "Alice" {
		t.Fatalf("unexpected answers: %+v", answers)
	}

	count, err := q.CountResponsesByFormID(ctx, form.ID)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 response, got %d", count)
	}
}

func TestDeleteExpiredMagicLinks(t *testing.T) {
	q := testQueries(t)
	ctx := context.Background()

	user := testUser(t, q, "expiredlinks@example.com")

	_, err := q.CreateMagicLink(ctx, CreateMagicLinkParams{
		UserID:    user.ID,
		Token:     "cleanup-expired-token",
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Minute), Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := q.DeleteExpiredMagicLinks(ctx); err != nil {
		t.Fatal(err)
	}

	_, err = q.GetMagicLinkByToken(ctx, "cleanup-expired-token")
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected expired link to be deleted, got %v", err)
	}
}
