package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/iliaonishchenko/aggreg8/internal/service"
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
		//{
		//	name:               "valid gauge metric with invalid method",
		//	requestURL:         "http://localhost:8080/update/gauge/temperature/23.5",
		//	requestMethod:      http.MethodGet,
		//	requestContentType: "text/plain",
		//	want: want{
		//		code:        http.StatusMethodNotAllowed,
		//		response:    "",
		//		contentType: "",
		//	},
		//},
		//{
		//	name:               "valid gauge metric with invalid content type",
		//	requestURL:         "http://localhost:8080/update/gauge/temperature/23.5",
		//	requestMethod:      http.MethodPost,
		//	requestContentType: "application/json",
		//	want: want{
		//		code:        http.StatusBadRequest,
		//		response:    "",
		//		contentType: "",
		//	},
		//},
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
			storage := service.NewMemStorage()
			handler := NewUpdateHandler(storage)

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
