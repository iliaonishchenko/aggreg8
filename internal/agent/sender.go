package agent

import (
	"fmt"
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

func (s *Sender) Send(metrics []*models.Metrics) {
	for _, v := range metrics {
		fmt.Printf("send metric: type: %s, name: %s\n", v.MType, v.ID)
		resp, err := s.sendMetric(v)
		if err != nil {
			fmt.Printf("error sending the metric: type: %s, name: %s, err: %v", v.MType, v.ID, err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("server returned error: %d\n", resp.StatusCode)
		}
		resp.Body.Close()
	}
}

func (s *Sender) sendMetric(metric *models.Metrics) (*http.Response, error) {
	metricURL := s.buildMetricURL(s.endpoint, metric)
	return s.client.Post(metricURL, "text/plain", nil)
}

func (s *Sender) buildMetricURL(baseURL string, metric *models.Metrics) string {
	var finalUrl string
	switch metric.MType {
	case models.Gauge:
		finalUrl = fmt.Sprintf("%s/update/%s/%s/%s",
			baseURL,
			url.PathEscape(metric.MType),
			url.PathEscape(metric.ID),
			url.PathEscape(fmt.Sprintf("%f", *metric.Value)),
		)
	case models.Counter:
		finalUrl = fmt.Sprintf("%s/update/%s/%s/%s",
			baseURL,
			url.PathEscape(metric.MType),
			url.PathEscape(metric.ID),
			url.PathEscape(strconv.FormatInt(*metric.Delta, 10)),
		)
	}
	return finalUrl
}
