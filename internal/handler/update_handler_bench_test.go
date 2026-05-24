package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/iliaonishchenko/aggreg8/internal/audit"
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/iliaonishchenko/aggreg8/internal/service"
	"github.com/iliaonishchenko/aggreg8/internal/service/memory"
)

func BenchmarkHandleUpdateJSON(b *testing.B) {
	logger.Initialize("error")
	storage := memory.NewMemStorage()
	h := NewUpdateHandler(service.NewRecorder(storage, audit.NewNotifier(10)))

	r := chi.NewRouter()
	r.Post("/update", h.HandleUpdateJSON)

	metric := models.Metrics{ID: "temperature", MType: models.Gauge, Value: ptrFloat64(23.5)}
	body, _ := json.Marshal(metric)

	for b.Loop() {
		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkHandleBatchUpdateJSON(b *testing.B) {
	logger.Initialize("error")

	for _, size := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			storage := memory.NewMemStorage()
			h := NewUpdateHandler(service.NewRecorder(storage, audit.NewNotifier(10)))

			r := chi.NewRouter()
			r.Post("/updates", h.HandleBatchUpdateJSON)

			metrics := make([]*models.Metrics, size)
			for i := range metrics {
				v := float64(i)
				metrics[i] = &models.Metrics{
					ID:    fmt.Sprintf("metric_%d", i),
					MType: models.Gauge,
					Value: &v,
				}
			}
			body, _ := json.Marshal(metrics)

			for b.Loop() {
				req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
			}
		})
	}
}
