package athletes

import (
	"log"
	"net/http"

	"orenkod/backend/internal/httpserver"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.List(r.Context())
	if err != nil {
		log.Printf("list athletes: %v", err)
		httpserver.WriteError(w, http.StatusInternalServerError, "Не удалось получить реестр")
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}
