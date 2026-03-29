package audit

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPObserver_Notify(t *testing.T) {
	var receivedEvent AuditEvent
	var receivedContentType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedEvent)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	observer := NewHTTPObserver(server.URL, http.Client{})

	event := AuditEvent{
		TS:        12345678,
		Metrics:   []string{"Alloc", "Frees"},
		IPAddress: "192.168.0.42",
	}

	err := observer.Notify(event)
	require.NoError(t, err)

	assert.Equal(t, "application/json", receivedContentType)
	assert.Equal(t, event, receivedEvent)
}

func TestHTTPObserver_Notify_ServerUnavailable(t *testing.T) {
	observer := NewHTTPObserver("http://127.0.0.1:1", http.Client{})

	err := observer.Notify(AuditEvent{TS: 1000, Metrics: []string{"Alloc"}, IPAddress: "127.0.0.1"})
	assert.Error(t, err)
}
