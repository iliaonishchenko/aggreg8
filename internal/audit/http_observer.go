package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// HTTPObserver отправляет аудит-события по HTTP POST в формате JSON.
type HTTPObserver struct {
	url    string
	client *http.Client
}

// NewHTTPObserver создаёт HTTPObserver, отправляющий события на указанный URL.
func NewHTTPObserver(url string, client *http.Client) *HTTPObserver {
	return &HTTPObserver{
		url:    url,
		client: client,
	}
}

// Notify отправляет событие на удалённый сервер аудита.
func (o *HTTPObserver) Notify(event AuditEvent) error {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(event); err != nil {
		return fmt.Errorf("could not marshal audit event to JSON: %w", err)
	}

	resp, err := o.client.Post(o.url, "application/json", buf)
	if err != nil {
		return fmt.Errorf("could not POST audit event: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
