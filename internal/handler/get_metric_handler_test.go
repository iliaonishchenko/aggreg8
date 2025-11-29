package handler

import (
	"github.com/go-chi/chi/v5"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/iliaonishchenko/aggreg8/internal/service"
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

func storageWithMetrics(metrics []*models.Metrics) service.MetricStorage {
	storage := service.NewMemStorage()
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
