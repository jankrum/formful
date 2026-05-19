package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/jankrum/formful/internal/database"
)

type nonceKey struct{}
type sessionKey struct{}

func GetNonce(ctx context.Context) string {
	v, _ := ctx.Value(nonceKey{}).(string)
	return v
}

func GetSession(ctx context.Context) *database.GetSessionByIDRow {
	v, _ := ctx.Value(sessionKey{}).(*database.GetSessionByIDRow)
	return v
}

func Nonce(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 16)
		rand.Read(b)
		nonce := base64.StdEncoding.EncodeToString(b)

		csp := fmt.Sprintf(
			"default-src 'self'; script-src 'nonce-%s'; style-src 'nonce-%s' 'self' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'",
			nonce, nonce,
		)
		w.Header().Set("Content-Security-Policy", csp)

		ctx := context.WithValue(r.Context(), nonceKey{}, nonce)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AuthSession(queries *database.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			session, err := queries.GetSessionByID(r.Context(), cookie.Value)
			if err != nil {
				http.SetCookie(w, &http.Cookie{Name: "session", MaxAge: -1, Path: "/"})
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			ctx := context.WithValue(r.Context(), sessionKey{}, &session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
