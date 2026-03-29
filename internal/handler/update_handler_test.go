package handler

import (
	"bytes"
	"encoding/json"
	"github.com/iliaonishchenko/aggreg8/internal/audit"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/iliaonishchenko/aggreg8/internal/service/memory"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestHandleUpdate(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name               string
		requestURL         string
		requestMethod      string
		requestContentType string
		want               want
	}{
		{
			name:               "valid gauge metric",
			requestURL:         "http://localhost:8080/update/gauge/temperature/23.5",
			requestMethod:      http.MethodPost,
			requestContentType: "text/plain",
			want: want{
				code:        200,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:               "valid counter metric",
			requestURL:         "http://localhost:8080/update/counter/temperature/23",
			requestMethod:      http.MethodPost,
			requestContentType: "text/plain",
			want: want{
				code:        200,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:               "gauge metric without name",
			requestURL:         "http://localhost:8080/update/gauge/23.5",
			requestMethod:      http.MethodPost,
			requestContentType: "text/plain",
			want: want{
				code:        http.StatusNotFound,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:               "invalid metric type",
			requestURL:         "http://localhost:8080/update/invalid_type/temperature/23.5",
			requestMethod:      http.MethodPost,
			requestContentType: "text/plain",
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				contentType: "",
			},
		},
		{
			name:               "invalid gauge value",
			requestURL:         "http://localhost:8080/update/gauge/temperature/hello",
			requestMethod:      http.MethodPost,
			requestContentType: "text/plain",
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				contentType: "",
			},
		},
		{
			name:               "invalid counter value",
			requestURL:         "http://localhost:8080/update/counter/temperature/hello",
			requestMethod:      http.MethodPost,
			requestContentType: "text/plain",
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				contentType: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.NewMemStorage()
			handler := NewUpdateHandler(storage, audit.NewNotifier())

			r := chi.NewRouter()
			r.Post("/update/{type}/{name}/{value}", handler.HandleUpdate)

			request := httptest.NewRequest(tt.requestMethod, tt.requestURL, nil)
			request.Header.Set("Content-Type", tt.requestContentType)

			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.code, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
		})
	}
}

func TestHandleUpdateJSON(t *testing.T) {
	tests := []struct {
		name          string
		requestMethod string
		requestBody   any
		responseCode  int
	}{
		{
			name:          "valid gauge metric",
			requestMethod: http.MethodPost,
			requestBody:   models.Metrics{ID: "temperature", MType: models.Gauge, Value: ptrFloat64(23.5)},
			responseCode:  http.StatusOK,
		},
		{
			name:          "valid counter metric",
			requestMethod: http.MethodPost,
			requestBody:   models.Metrics{ID: "requests", MType: models.Counter, Delta: ptrInt64(10)},
			responseCode:  http.StatusOK,
		},
		{
			name:          "missing metric ID",
			requestMethod: http.MethodPost,
			requestBody:   map[string]interface{}{"value": ptrFloat64(23.5)},
			responseCode:  http.StatusBadRequest,
		},
		{
			name:          "missing metric type",
			requestMethod: http.MethodPost,
			requestBody:   map[string]interface{}{"id": "temperature", "value": ptrFloat64(23.5)},
			responseCode:  http.StatusBadRequest,
		},
		{
			name:          "invalid gauge metric",
			requestMethod: http.MethodPost,
			requestBody:   map[string]interface{}{"id": "temperature", "type": models.Gauge},
			responseCode:  http.StatusBadRequest,
		},
		{
			name:          "invalid counter metric",
			requestMethod: http.MethodPost,
			requestBody:   map[string]interface{}{"id": "requests", "type": models.Counter},
			responseCode:  http.StatusBadRequest,
		},
		{
			name:          "non-existing metric",
			requestMethod: http.MethodPost,
			requestBody:   map[string]interface{}{"id": "requests", "type": "len"},
			responseCode:  http.StatusBadRequest,
		},
		{
			name:          "invalid method",
			requestMethod: http.MethodGet,
			requestBody:   nil,
			responseCode:  http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestURL := "http://localhost:8080/update"
			storage := memory.NewMemStorage()
			handler := NewUpdateHandler(storage, audit.NewNotifier())

			r := chi.NewRouter()
			r.Post("/update", handler.HandleUpdateJSON)

			buf := new(bytes.Buffer)
			enc := json.NewEncoder(buf)
			if err := enc.Encode(tt.requestBody); err != nil {
				t.Fatalf("failed to encode request body: %v", err)
			}
			request := httptest.NewRequest(tt.requestMethod, requestURL, buf)

			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.responseCode, result.StatusCode)
		})
	}
}

func TestHandleBatchUpdateJSON(t *testing.T) {
	tests := []struct {
		name          string
		requestMethod string
		requestBody   any
		responseCode  int
	}{
		{
			name:          "valid batch with multiple metrics",
			requestMethod: http.MethodPost,
			requestBody: []*models.Metrics{
				{ID: "temperature", MType: models.Gauge, Value: ptrFloat64(23.5)},
				{ID: "requests", MType: models.Counter, Delta: ptrInt64(10)},
				{ID: "cpu", MType: models.Gauge, Value: ptrFloat64(75.0)},
			},
			responseCode: http.StatusOK,
		},
		{
			name:          "valid batch with single metric",
			requestMethod: http.MethodPost,
			requestBody: []*models.Metrics{
				{ID: "temperature", MType: models.Gauge, Value: ptrFloat64(23.5)},
			},
			responseCode: http.StatusOK,
		},
		{
			name:          "empty batch",
			requestMethod: http.MethodPost,
			requestBody:   []*models.Metrics{},
			responseCode:  http.StatusOK,
		},
		{
			name:          "batch with missing metric ID",
			requestMethod: http.MethodPost,
			requestBody: []*models.Metrics{
				{ID: "temperature", MType: models.Gauge, Value: ptrFloat64(23.5)},
				{MType: models.Counter, Delta: ptrInt64(10)},
			},
			responseCode: http.StatusBadRequest,
		},
		{
			name:          "batch with missing metric type",
			requestMethod: http.MethodPost,
			requestBody: []*models.Metrics{
				{ID: "temperature", MType: models.Gauge, Value: ptrFloat64(23.5)},
				{ID: "requests", Delta: ptrInt64(10)},
			},
			responseCode: http.StatusBadRequest,
		},
		{
			name:          "batch with missing gauge value",
			requestMethod: http.MethodPost,
			requestBody: []*models.Metrics{
				{ID: "temperature", MType: models.Gauge, Value: ptrFloat64(23.5)},
				{ID: "cpu", MType: models.Gauge},
			},
			responseCode: http.StatusBadRequest,
		},
		{
			name:          "batch with missing counter delta",
			requestMethod: http.MethodPost,
			requestBody: []*models.Metrics{
				{ID: "temperature", MType: models.Gauge, Value: ptrFloat64(23.5)},
				{ID: "requests", MType: models.Counter},
			},
			responseCode: http.StatusBadRequest,
		},
		{
			name:          "batch with invalid metric type",
			requestMethod: http.MethodPost,
			requestBody: []*models.Metrics{
				{ID: "temperature", MType: models.Gauge, Value: ptrFloat64(23.5)},
				{ID: "invalid", MType: "unknown", Value: ptrFloat64(1.0)},
			},
			responseCode: http.StatusBadRequest,
		},
		{
			name:          "invalid JSON body",
			requestMethod: http.MethodPost,
			requestBody:   "invalid json",
			responseCode:  http.StatusBadRequest,
		},
		{
			name:          "invalid method",
			requestMethod: http.MethodGet,
			requestBody:   nil,
			responseCode:  http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestURL := "http://localhost:8080/updates"
			storage := memory.NewMemStorage()
			handler := NewUpdateHandler(storage, audit.NewNotifier())

			r := chi.NewRouter()
			r.Post("/updates", handler.HandleBatchUpdateJSON)

			buf := new(bytes.Buffer)
			enc := json.NewEncoder(buf)
			if err := enc.Encode(tt.requestBody); err != nil {
				t.Fatalf("failed to encode request body: %v", err)
			}
			request := httptest.NewRequest(tt.requestMethod, requestURL, buf)

			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.responseCode, result.StatusCode)
		})
	}
}

func ptrFloat64(v float64) *float64 {
	return &v
}

func ptrInt64(v int64) *int64 {
	return &v
}
