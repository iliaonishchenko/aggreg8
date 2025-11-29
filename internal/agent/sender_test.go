package agent

import (
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildMetricURL(t *testing.T) {
	baseURL := "localhost:8080"

	tests := []struct {
		name        string
		metricModel *models.Metrics
		expectedURL string
		err         error
	}{
		{
			name: "build gauge metric",
			metricModel: &models.Metrics{
				ID:    "temperature",
				MType: models.Gauge,
				Value: float64Ptr(23.5),
			},
			expectedURL: "http://localhost:8080/update/gauge/temperature/23.5",
			err:         nil,
		},
		{
			name: "build counter metric",
			metricModel: &models.Metrics{
				ID:    "requests",
				MType: models.Counter,
				Delta: int64Ptr(42),
			},
			expectedURL: "http://localhost:8080/update/counter/requests/42",
			err:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := NewSender(baseURL)
			actualURL := sender.buildMetricURL(baseURL, tt.metricModel)
			assert.Equal(t, tt.expectedURL, actualURL)
		})
	}
}
