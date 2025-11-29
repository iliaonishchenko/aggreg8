package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleAllMetrics(t *testing.T) {
	type want struct {
		code            int
		contentType     string
		containsStrings []string
	}
	tests := []struct {
		name          string
		storedMetrics []*models.Metrics
		want          want
	}{
		{
			name:          "empty storage returns empty table",
			storedMetrics: []*models.Metrics{},
			want: want{
				code:        http.StatusOK,
				contentType: "text/html; charset=utf-8",
				containsStrings: []string{
					"<html>",
					"<h1>All Metrics</h1>",
					"<table",
					"</table>",
				},
			},
		},
		{
			name: "storage with gauge metric",
			storedMetrics: []*models.Metrics{
				{ID: "temperature", MType: models.Gauge, Value: floatPtr(23.5)},
			},
			want: want{
				code:        http.StatusOK,
				contentType: "text/html; charset=utf-8",
				containsStrings: []string{
					"<td>temperature</td>",
					"<td>gauge</td>",
					"<td>23.5</td>",
				},
			},
		},
		{
			name: "storage with counter metric",
			storedMetrics: []*models.Metrics{
				{ID: "requests", MType: models.Counter, Value: floatPtr(42)},
			},
			want: want{
				code:        http.StatusOK,
				contentType: "text/html; charset=utf-8",
				containsStrings: []string{
					"<td>requests</td>",
					"<td>counter</td>",
					"<td>42</td>",
				},
			},
		},
		{
			name: "storage with multiple metrics",
			storedMetrics: []*models.Metrics{
				{ID: "temperature", MType: models.Gauge, Value: floatPtr(23.5)},
				{ID: "requests", MType: models.Counter, Value: floatPtr(100)},
			},
			want: want{
				code:        http.StatusOK,
				contentType: "text/html; charset=utf-8",
				containsStrings: []string{
					"<td>temperature</td>",
					"<td>requests</td>",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := storageWithMetrics(tt.storedMetrics)
			handler := NewAllMetricsHandler(storage)

			r := chi.NewRouter()
			r.Get("/", handler.HandleAll)

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			result := w.Result()
			defer result.Body.Close()

			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.want.code, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			for _, s := range tt.want.containsStrings {
				assert.True(t, strings.Contains(string(body), s), "response should contain: %s", s)
			}
		})
	}
}
