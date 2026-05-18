package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jankrum/formful/cmd/web"
	webauth "github.com/jankrum/formful/cmd/web/auth"
	"github.com/jankrum/formful/cmd/web/dashboard"
	"github.com/jankrum/formful/internal/flash"
	"github.com/jankrum/formful/internal/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(methodOverride)
	r.Use(middleware.Nonce)
	r.Use(flash.Middleware)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	fileServer := http.FileServer(http.FS(web.Files))
	r.Handle("/assets/*", fileServer)

	r.Get("/health", s.healthHandler)

	authHandler := webauth.NewHandler(s.queries, s.mailer)
	r.Get("/login", authHandler.LoginPage)
	r.Post("/login", authHandler.PostLogin)
	r.Get("/link-sent", authHandler.LinkSentPage)
	r.Get("/auth/verify", authHandler.VerifyMagicLink)
	r.Post("/logout", authHandler.Logout)

	dashboardHandler := dashboard.NewHandler(s.queries)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthSession(s.queries))
		r.Get("/", dashboardHandler.DashboardPage)
	})

	return r
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.db.Health())
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

func methodOverride(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			r.ParseForm()
			if m := r.FormValue("_method"); m != "" {
				r.Method = strings.ToUpper(m)
			}
		}
		next.ServeHTTP(w, r)
	})
}
