package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jankrum/formful/cmd/web"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(methodOverride)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	fileServer := http.FileServer(http.FS(web.Files))
	r.Handle("/assets/*", fileServer)

	r.Get("/", templ.Handler(web.Index()).ServeHTTP)
	r.Get("/health", s.healthHandler)

	return r
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.db.Health())
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

// methodOverride reads _method from POST form data and overrides r.Method.
// This enables DELETE/PUT/PATCH from HTML forms in no-JS environments.
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
