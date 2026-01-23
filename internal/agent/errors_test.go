package agent

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRetriable(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
		want       bool
	}{
		{
			name:       "connection error - retriable",
			err:        errors.New("connection refused"),
			statusCode: 0,
			want:       true,
		},
		{
			name:       "timeout error - retriable",
			err:        errors.New("timeout"),
			statusCode: 0,
			want:       true,
		},
		{
			name:       "500 internal server error - retriable",
			err:        nil,
			statusCode: 500,
			want:       true,
		},
		{
			name:       "502 bad gateway - retriable",
			err:        nil,
			statusCode: 502,
			want:       true,
		},
		{
			name:       "503 service unavailable - retriable",
			err:        nil,
			statusCode: 503,
			want:       true,
		},
		{
			name:       "504 gateway timeout - retriable",
			err:        nil,
			statusCode: 504,
			want:       true,
		},
		{
			name:       "200 OK - not retriable",
			err:        nil,
			statusCode: 200,
			want:       false,
		},
		{
			name:       "400 bad request - not retriable",
			err:        nil,
			statusCode: 400,
			want:       false,
		},
		{
			name:       "401 unauthorized - not retriable",
			err:        nil,
			statusCode: 401,
			want:       false,
		},
		{
			name:       "404 not found - not retriable",
			err:        nil,
			statusCode: 404,
			want:       false,
		},
		{
			name:       "409 conflict - not retriable",
			err:        nil,
			statusCode: 409,
			want:       false,
		},
	}

	classifier := NewAgentErrorClassifier()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *http.Response
			if tt.err == nil {
				resp = &http.Response{StatusCode: tt.statusCode}
			}

			got := classifier.isRetriable(tt.err, resp)
			assert.Equal(t, tt.want, got)
		})
	}
}
