package handler

import (
	"github.com/iliaonishchenko/aggreg8/internal/repository"
	"net/http"
)

type PingHandler struct {
	db *repository.MetricsRepository
}

func NewPingHandler(db *repository.MetricsRepository) *PingHandler {
	return &PingHandler{db: db}
}

func (h *PingHandler) HandlePing(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := h.db.Ping(); err != nil {
		http.Error(w, "MetricsRepository connection error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
