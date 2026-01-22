package handler

import (
	"bytes"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/iliaonishchenko/aggreg8/internal/service"
	"github.com/iliaonishchenko/aggreg8/internal/service/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleGetMetric(t *testing.T) {
	type want struct {
		code     int
		response string
	}
	tests := []struct {
		name               string
		requestURL         string
		requestMethod      string
		requestContentType string
		storedMetrics      []*models.Metrics
		want               want
	}{
		{
			name:               "valid get gauge metric",
			requestURL:         "http://localhost:8080/value/gauge/temperature",
			requestMethod:      "GET",
			requestContentType: "text/plain",
			storedMetrics: []*models.Metrics{
				{ID: "temperature", MType: models.Gauge, Value: floatPtr(23.5)},
			},
			want: want{
				code:     200,
				response: "23.5",
			},
		},
		{
			name:               "valid get counter metric",
			requestURL:         "http://localhost:8080/value/counter/requests",
			requestMethod:      "GET",
			requestContentType: "text/plain",
			storedMetrics: []*models.Metrics{
				{ID: "requests", MType: models.Counter, Delta: intPtr(42)},
			},
			want: want{
				code:     200,
				response: "42",
			},
		},
		{
			name:               "get unknown metric",
			requestURL:         "http://localhost:8080/value/gauge/unknown_metric",
			requestMethod:      "GET",
			requestContentType: "text/plain",
			storedMetrics: []*models.Metrics{
				{ID: "temperature", MType: models.Gauge, Value: floatPtr(23.5)},
			},
			want: want{
				code:     http.StatusNotFound,
				response: "unknown metric: unknown_metric\n",
			},
		},
		{
			name:               "get metric with unknown type",
			requestURL:         "http://localhost:8080/value/unknown_type/temperature",
			requestMethod:      "GET",
			requestContentType: "text/plain",
			storedMetrics: []*models.Metrics{
				{ID: "temperature", MType: models.Gauge, Value: floatPtr(23.5)},
			},
			want: want{
				code:     http.StatusNotFound,
				response: "unknown metric type: unknown_type\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := storageWithMetrics(tt.storedMetrics)
			handler := NewGetMetricHandler(storage)

			r := chi.NewRouter()
			r.Get("/value/{type}/{name}", handler.HandleGetMetric)

			request := httptest.NewRequest(tt.requestMethod, tt.requestURL, nil)
			request.Header.Add("Content-Type", tt.requestContentType)

			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			result := w.Result()
			defer result.Body.Close()

			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.want.code, result.StatusCode)
			assert.Equal(t, tt.want.response, string(body))
		})
	}
}

func TestHandleGetMetricJSON(t *testing.T) {
	validGaugeMetric := models.Metrics{ID: "temperature", MType: models.Gauge, Value: floatPtr(23.5)}
	validCounterMetric := models.Metrics{ID: "requests", MType: models.Counter, Delta: intPtr(42)}

	type want struct {
		code         int
		responseBody *models.Metrics
	}

	tests := []struct {
		name          string
		requestMethod string
		requestBody   any
		want          want
	}{
		{
			name:          "valid get gauge metric",
			requestMethod: http.MethodPost,
			requestBody:   models.Metrics{ID: "temperature", MType: models.Gauge},
			want: want{
				code:         200,
				responseBody: &validGaugeMetric,
			},
		},
		{
			name:          "valid get counter metric",
			requestMethod: http.MethodPost,
			requestBody:   models.Metrics{ID: "requests", MType: models.Counter},
			want: want{
				code:         200,
				responseBody: &validCounterMetric,
			},
		},
		{
			name:          "incorrect metric - missing ID",
			requestMethod: http.MethodPost,
			requestBody:   models.Metrics{MType: models.Gauge},
			want: want{
				code:         http.StatusBadRequest,
				responseBody: nil,
			},
		},
		{
			name:          "incorrect metric - missing type",
			requestMethod: http.MethodPost,
			requestBody:   models.Metrics{ID: "temperature"},
			want: want{
				code:         http.StatusBadRequest,
				responseBody: nil,
			},
		},
		{
			name:          "non-existing metric type",
			requestMethod: http.MethodPost,
			requestBody:   models.Metrics{ID: "requests", MType: "len"},
			want: want{
				code:         http.StatusBadRequest,
				responseBody: nil,
			},
		},
		{
			name:          "non existing metric",
			requestMethod: http.MethodPost,
			requestBody:   models.Metrics{ID: "non-existing-metric", MType: models.Gauge},
			want: want{
				code:         http.StatusNotFound,
				responseBody: &models.Metrics{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storedMetrics := []*models.Metrics{&validGaugeMetric, &validCounterMetric}
			storage := storageWithMetrics(storedMetrics)
			handler := NewGetMetricHandler(storage)

			r := chi.NewRouter()
			r.Post("/value/", handler.HandleGetMetricJSON)

			buf := new(bytes.Buffer)
			enc := json.NewEncoder(buf)
			if err := enc.Encode(tt.requestBody); err != nil {
				t.Fatalf("failed to encode request body: %v", err)
			}

			request := httptest.NewRequest(tt.requestMethod, "http://localhost:8080/value/", buf)
			request.Header.Add("Content-Type", "application/json")

			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			result := w.Result()
			defer result.Body.Close()

			if tt.want.responseBody != nil {
				var actualBody models.Metrics
				dec := json.NewDecoder(result.Body)
				err := dec.Decode(&actualBody)
				require.NoError(t, err)
				assert.Equal(t, tt.want.responseBody, &actualBody)
			} else {
				body, err := io.ReadAll(result.Body)
				require.NoError(t, err)
				assert.Empty(t, body)
			}
			assert.Equal(t, tt.want.code, result.StatusCode)
		})
	}
}

func storageWithMetrics(metrics []*models.Metrics) service.MetricStorage {
	storage := memory.NewMemStorage()
	for _, metric := range metrics {
		storage.UpdateMetric(metric)
	}
	return storage
}

func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int64) *int64 {
	return &i
}
