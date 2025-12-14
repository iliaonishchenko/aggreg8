package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"net/http"
	"net/url"
	"strconv"
)

type Sender struct {
	endpoint string
	client   http.Client
}

func NewSender(endpoint string) *Sender {
	return &Sender{
		endpoint: endpoint,
		client:   http.Client{},
	}
}

func (s *Sender) SendJSON(metric *models.Metrics) error {
	uri := fmt.Sprintf("http://%s/update", s.endpoint)
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(metric); err != nil {
		logger.Log.Error("error encoding metric to JSON", logger.Err(err))
	}

	resp, err := s.client.Post(uri, "application/json", buf)

	if err != nil {
		logger.Log.Error("error sending metric to agent", logger.Err(err))
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Log.Info("error sending metric to agent", logger.Err(err))
		return err
	}

	return nil
}

func (s *Sender) Send(metric *models.Metrics) error {
	resp, err := s.sendMetric(metric)
	if err != nil {
		fmt.Printf("error sending the metric: type: %s, name: %s, err: %v", metric.MType, metric.ID, err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("server returned error: %d\n", resp.StatusCode)
	}
	resp.Body.Close()
	return nil
}

func (s *Sender) sendMetric(metric *models.Metrics) (*http.Response, error) {
	metricURL := s.buildMetricURL(s.endpoint, metric)

	return s.client.Post(metricURL, "text/plain", nil)
}

func (s *Sender) buildMetricURL(baseURL string, metric *models.Metrics) string {
	var metricValue string

	switch metric.MType {
	case models.Gauge:
		metricValue = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	case models.Counter:
		metricValue = strconv.FormatInt(*metric.Delta, 10)
	}

	return fmt.Sprintf("http://%s/update/%s/%s/%s",
		baseURL,
		url.PathEscape(metric.MType),
		url.PathEscape(metric.ID),
		url.PathEscape(metricValue),
	)
}
