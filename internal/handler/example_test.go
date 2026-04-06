package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/iliaonishchenko/aggreg8/internal/audit"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/iliaonishchenko/aggreg8/internal/service/memory"
)

func ExampleUpdateHandler_HandleUpdate() {
	storage := memory.NewMemStorage()
	notifier := audit.NewNotifier(10)
	h := NewUpdateHandler(storage, notifier)

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h.HandleUpdate)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/temperature/23.5", nil)
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output:
	// 200
}

func ExampleUpdateHandler_HandleUpdateJSON() {
	storage := memory.NewMemStorage()
	notifier := audit.NewNotifier(10)
	h := NewUpdateHandler(storage, notifier)

	r := chi.NewRouter()
	r.Post("/update", h.HandleUpdateJSON)

	value := 36.6
	metric := models.Metrics{
		ID:    "temperature",
		MType: models.Gauge,
		Value: &value,
	}

	body, _ := json.Marshal(metric)
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output:
	// 200
}

func ExampleUpdateHandler_HandleBatchUpdateJSON() {
	storage := memory.NewMemStorage()
	notifier := audit.NewNotifier(10)
	h := NewUpdateHandler(storage, notifier)

	r := chi.NewRouter()
	r.Post("/updates", h.HandleBatchUpdateJSON)

	value := 36.6
	var delta int64 = 10
	metrics := []*models.Metrics{
		{ID: "temperature", MType: models.Gauge, Value: &value},
		{ID: "requests", MType: models.Counter, Delta: &delta},
	}

	body, _ := json.Marshal(metrics)
	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output:
	// 200
}

func ExampleGetMetricHandler_HandleGetMetric() {
	storage := memory.NewMemStorage()
	value := 23.5
	storage.UpdateMetric(&models.Metrics{ID: "temperature", MType: models.Gauge, Value: &value})

	h := NewGetMetricHandler(storage)

	r := chi.NewRouter()
	r.Get("/value/{type}/{name}", h.HandleGetMetric)

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/temperature", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	body, _ := io.ReadAll(w.Body)
	fmt.Println(w.Code)
	fmt.Println(string(body))
	// Output:
	// 200
	// 23.5
}

func ExampleGetMetricHandler_HandleGetMetricJSON() {
	storage := memory.NewMemStorage()
	value := 42.0
	storage.UpdateMetric(&models.Metrics{ID: "cpu", MType: models.Gauge, Value: &value})

	h := NewGetMetricHandler(storage)

	r := chi.NewRouter()
	r.Post("/value", h.HandleGetMetricJSON)

	reqBody := models.Metrics{ID: "cpu", MType: models.Gauge}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	var result models.Metrics
	json.NewDecoder(w.Body).Decode(&result)

	fmt.Println(w.Code)
	fmt.Println(result.ID)
	fmt.Println(*result.Value)
	// Output:
	// 200
	// cpu
	// 42
}

func ExampleAllMetricsHandler_HandleAll() {
	storage := memory.NewMemStorage()
	value := 23.5
	storage.UpdateMetric(&models.Metrics{ID: "temperature", MType: models.Gauge, Value: &value})

	h := NewAllMetricsHandler(storage)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.HandleAll(w, req)

	fmt.Println(w.Code)
	fmt.Println(w.Header().Get("Content-Type"))
	// Output:
	// 200
	// text/html; charset=utf-8
}
