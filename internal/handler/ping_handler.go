package handler

import (
	"net/http"
)

// Repository определяет интерфейс для проверки доступности хранилища данных.
type Repository interface {
	Ping() error
}

// PingHandler обрабатывает HTTP-запрос проверки доступности базы данных: GET /ping.
type PingHandler struct {
	db Repository
}

// NewPingHandler создаёт новый PingHandler с указанным репозиторием.
func NewPingHandler(db Repository) *PingHandler {
	return &PingHandler{db: db}
}

// HandlePing проверяет соединение с базой данных и возвращает 200 OK при успехе.
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
