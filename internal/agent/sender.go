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
	"time"
)

type DataSignature interface {
	Sign(src []byte) string
}

type Sender struct {
	endpoint   string
	client     http.Client
	classifier *AgentErrorClassifier
	signature  DataSignature
}

func NewSender(endpoint string, classifier *AgentErrorClassifier, signature DataSignature) *Sender {
	return &Sender{
		endpoint:   endpoint,
		client:     http.Client{},
		classifier: classifier,
		signature:  signature,
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
	req, err := s.buildRequest(metrics...)
	if err != nil {
		return fmt.Errorf("error building request to agent %w", err)
	}

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

func (s *Sender) SendJSONWithRetries(metrics ...*models.Metrics) error {
	req, err := s.buildRequest(metrics...)
	if err != nil {
		return fmt.Errorf("error building request to agent %w", err)
	}

	return s.sendWithRetries(req)
}

func (s *Sender) buildRequest(metrics ...*models.Metrics) (*http.Request, error) {
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
		return nil, err
	}

	compressedBuf, err := compressData(metricBuf.Bytes())
	if err != nil {
		logger.Log.Error("error compressing metric data", logger.Err(err))
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, compressedBuf)
	if err != nil {
		logger.Log.Error("error creating request to agent", logger.Err(err))
		return nil, err
	}

	signature, signed := s.getSignature(metricBuf.Bytes())
	if signed {
		req.Header.Set("HashSHA256", signature)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	return req, nil
}

func (s *Sender) doSingleRequest(req *http.Request) (*http.Response, error) {
	resp, err := s.client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	return resp, err
}

func (s *Sender) sendWithRetries(req *http.Request) error {
	const (
		maxAttempts = 4
		deltaDelay  = 2 * time.Second
	)
	currDelay := 1 * time.Second
	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt != 0 {
			time.Sleep(currDelay)
			currDelay += deltaDelay
		}

		resp, err := s.doSingleRequest(req)
		if resp != nil {
			defer resp.Body.Close()
		}

		if err == nil && resp.StatusCode == http.StatusOK {
			return nil
		}

		if !s.classifier.isRetriable(err, resp) {
			if err != nil {
				return err
			}
			return fmt.Errorf("server returned non-retriable status: %d", resp.StatusCode)
		}

		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("server returned status: %d", resp.StatusCode)
		}
	}
	return fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}

func (s *Sender) getSignature(src []byte) (string, bool) {
	if s.signature == nil {
		return "", false
	}
	return s.signature.Sign(src), true
}
