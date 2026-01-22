package agent

import (
	"bytes"
	"compress/gzip"
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

func encodeJSON(v interface{}) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf, nil
}

func compressData(data []byte) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	w := gzip.NewWriter(buf)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	err := w.Close()
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (s *Sender) SendJSON(metrics ...*models.Metrics) error {
	var uri string
	var metricBuf *bytes.Buffer
	var err error

	if len(metrics) == 1 {
		uri = fmt.Sprintf("http://%s/update", s.endpoint)
		metric := metrics[0]
		metricBuf, err = encodeJSON(metric)
	} else {
		uri = fmt.Sprintf("http://%s/updates", s.endpoint)
		metricBuf, err = encodeJSON(metrics)
	}

	if err != nil {
		logger.Log.Error("error encoding metric to JSON", logger.Err(err))
		return err
	}

	compressedBuf, err := compressData(metricBuf.Bytes())
	if err != nil {
		logger.Log.Error("error compressing metric data", logger.Err(err))
		return err
	}

	req, err := http.NewRequest("POST", uri, compressedBuf)
	if err != nil {
		logger.Log.Error("error creating request to agent", logger.Err(err))
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := s.client.Do(req)

	if err != nil {
		logger.Log.Error("error sending metric to agent", logger.Err(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("server returned non-OK status: %d", resp.StatusCode)
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
