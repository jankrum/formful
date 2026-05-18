package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/jankrum/formful/internal/database"
	"github.com/jankrum/formful/internal/email"
)

type Server struct {
	port    int
	db      database.Service
	queries *database.Queries
	mailer  email.Sender
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	db := database.NewService()
	queries := database.New(db.Pool())
	mailer := email.NewResendSender()

	s := &Server{
		port:    port,
		db:      db,
		queries: queries,
		mailer:  mailer,
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}
