package dashboard

import (
	"net/http"

	"github.com/jankrum/formful/internal/database"
)

type Handler struct {
	queries *database.Queries
}

func NewHandler(queries *database.Queries) *Handler {
	return &Handler{queries: queries}
}

func (h *Handler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	dashboardPage().Render(r.Context(), w)
}
