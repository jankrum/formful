package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/mail"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	internalauth "github.com/jankrum/formful/internal/auth"
	"github.com/jankrum/formful/internal/database"
	"github.com/jankrum/formful/internal/email"
	"github.com/jankrum/formful/internal/flash"
)

type Handler struct {
	queries *database.Queries
	mailer  email.Sender
}

func NewHandler(queries *database.Queries, mailer email.Sender) *Handler {
	return &Handler{queries: queries, mailer: mailer}
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// ?expired=1 comes from VerifyMagicLink — flash cookie unreliable on cross-site redirects
	if flash.GetEmailError(ctx) == "" && r.URL.Query().Get("expired") == "1" {
		ctx = flash.WithEmailError(ctx, "Invalid or expired login link.")
	}
	loginPage().Render(ctx, w)
}

func (h *Handler) PostLogin(w http.ResponseWriter, r *http.Request) {
	emailAddr := r.FormValue("email")
	if _, err := mail.ParseAddress(emailAddr); err != nil {
		flash.SetEmailError(w, "Invalid email address.")
		flash.SetEmailValue(w, emailAddr)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	token, err := internalauth.GenerateToken()
	if err != nil {
		slog.Error("generate token", "err", err)
		flash.Set(w, "Something went wrong. Please try again.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	verifyURL := fmt.Sprintf("%s://%s/auth/verify?token=%s", scheme, r.Host, token)

	if err := h.mailer.SendMagicLink(emailAddr, verifyURL); err != nil {
		slog.Error("send magic link", "err", err)
		flash.SetEmailError(w, "Failed to send email. Please try again.")
		flash.SetEmailValue(w, emailAddr)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := h.queries.UpsertUser(r.Context(), emailAddr)
	if err != nil {
		slog.Error("upsert user", "err", err)
		flash.Set(w, "Something went wrong. Please try again.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	_, err = h.queries.CreateMagicLink(r.Context(), database.CreateMagicLinkParams{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(15 * time.Minute), Valid: true},
	})
	if err != nil {
		slog.Error("create magic link", "err", err)
		flash.Set(w, "Something went wrong. Please try again.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/link-sent", http.StatusSeeOther)
}

func (h *Handler) LinkSentPage(w http.ResponseWriter, r *http.Request) {
	linkSentPage().Render(r.Context(), w)
}

func (h *Handler) VerifyMagicLink(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	magicLink, err := h.queries.GetMagicLinkByToken(r.Context(), token)
	if err != nil {
		http.Redirect(w, r, "/login?expired=1", http.StatusSeeOther)
		return
	}

	sessionID, err := internalauth.GenerateToken()
	if err != nil {
		slog.Error("generate session id", "err", err)
		flash.Set(w, "Something went wrong. Please try again.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	csrfToken, err := internalauth.GenerateToken()
	if err != nil {
		slog.Error("generate csrf token", "err", err)
		flash.Set(w, "Something went wrong. Please try again.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	_, err = h.queries.CreateSession(r.Context(), database.CreateSessionParams{
		ID:        sessionID,
		UserID:    magicLink.UserID,
		CsrfToken: csrfToken,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true},
	})
	if err != nil {
		slog.Error("create session", "err", err)
		flash.Set(w, "Something went wrong. Please try again.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := h.queries.DeleteMagicLink(r.Context(), magicLink.ID); err != nil {
		slog.Error("delete magic link", "err", err)
	}

	secure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(30 * 24 * time.Hour / time.Second),
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session"); err == nil {
		h.queries.DeleteSession(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
