package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type HTTPObserver struct {
	url    string
	client http.Client
}

func NewHTTPObserver(url string, client http.Client) *HTTPObserver {
	return &HTTPObserver{
		url:    url,
		client: client,
	}
}

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
