package router

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithCompression(t *testing.T) {
	tests := []struct {
		name                  string
		contentEncoding       string
		requestBody           string
		compressRequestBody   bool
		acceptEncoding        string
		responseStatus        int
		responseContentType   string
		responseBody          string
		expectCompressed      bool
		expectContentEncoding string
		expectRequestBody     string
	}{
		{
			name:                  "request with Content-Encoding: gzip -> decompress",
			contentEncoding:       "gzip",
			requestBody:           `{"data":"test"}`,
			compressRequestBody:   true,
			acceptEncoding:        "",
			responseStatus:        http.StatusOK,
			responseContentType:   "text/plain",
			responseBody:          "ok",
			expectCompressed:      false,
			expectContentEncoding: "",
			expectRequestBody:     `{"data":"test"}`,
		},
		{
			name:                  "request without Content-Encoding: gzip -> no decompress",
			contentEncoding:       "",
			requestBody:           `{"data":"test"}`,
			compressRequestBody:   false,
			acceptEncoding:        "",
			responseStatus:        http.StatusOK,
			responseContentType:   "text/plain",
			responseBody:          "ok",
			expectCompressed:      false,
			expectContentEncoding: "",
			expectRequestBody:     `{"data":"test"}`,
		},
		{
			name:                  "request with Accept-Encoding gzip -> compressed JSON response",
			contentEncoding:       "",
			requestBody:           "",
			compressRequestBody:   false,
			acceptEncoding:        "gzip",
			responseStatus:        http.StatusOK,
			responseContentType:   "application/json",
			responseBody:          `{"message":"hello"}`,
			expectCompressed:      true,
			expectContentEncoding: "gzip",
			expectRequestBody:     "",
		},
		{
			name:                  "request with Accept-Encoding gzip + non compressed text/plain response",
			contentEncoding:       "",
			requestBody:           "",
			compressRequestBody:   false,
			acceptEncoding:        "gzip",
			responseStatus:        http.StatusOK,
			responseContentType:   "text/plain",
			responseBody:          "plain text",
			expectCompressed:      false,
			expectContentEncoding: "",
			expectRequestBody:     "",
		},
		{
			name:                  "request without Accept-Encoding -> non-compressed response",
			contentEncoding:       "",
			requestBody:           "",
			compressRequestBody:   false,
			acceptEncoding:        "",
			responseStatus:        http.StatusOK,
			responseContentType:   "application/json",
			responseBody:          `{"message":"hello"}`,
			expectCompressed:      false,
			expectContentEncoding: "",
			expectRequestBody:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedBody string
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.requestBody != "" {
					body, _ := io.ReadAll(r.Body)
					receivedBody = string(body)
				}

				w.Header().Set("Content-Type", tt.responseContentType)
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			})

			wrapped := WithCompression(handler)

			var reqBody io.Reader
			if tt.requestBody != "" {
				if tt.compressRequestBody {
					var buf bytes.Buffer
					gzw := gzip.NewWriter(&buf)
					gzw.Write([]byte(tt.requestBody))
					gzw.Close()
					reqBody = &buf
				} else {
					reqBody = bytes.NewBufferString(tt.requestBody)
				}
			}

			req := httptest.NewRequest("POST", "/test", reqBody)
			if tt.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}

			rec := httptest.NewRecorder()
			wrapped.ServeHTTP(rec, req)

			if tt.expectRequestBody != "" && receivedBody != tt.expectRequestBody {
				t.Errorf("Handler received body: got %q, want %q", receivedBody, tt.expectRequestBody)
			}

			gotEncoding := rec.Header().Get("Content-Encoding")
			if gotEncoding != tt.expectContentEncoding {
				t.Errorf("Content-Encoding: got %q, want %q", gotEncoding, tt.expectContentEncoding)
			}

			var actualBody string
			if tt.expectCompressed {
				reader, err := gzip.NewReader(rec.Body)
				if err != nil {
					t.Fatalf("Failed to create gzip reader: %v", err)
				}
				decompressed, err := io.ReadAll(reader)
				if err != nil {
					t.Fatalf("Failed to read decompressed data: %v", err)
				}
				actualBody = string(decompressed)
			} else {
				actualBody = rec.Body.String()
			}

			if actualBody != tt.responseBody {
				t.Errorf("Response body: got %q, want %q", actualBody, tt.responseBody)
			}
		})
	}
}
