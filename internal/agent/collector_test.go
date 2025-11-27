package agent

import (
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCollect(t *testing.T) {
	tests := []struct {
		name               string
		pollValue1         float64
		pollValue2         float64
		expectedMetricName string
	}{
		{
			name:               "collect metric test",
			pollValue1:         0,
			pollValue2:         1,
			expectedMetricName: "Alloc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewCollector()
			collector.Collect()

			metrics := collector.GetMetrics()
			assert.Contains(t, getNames(metrics), tt.expectedMetricName)
			assert.Equal(t, tt.pollValue1, *getMetricByName("PollCount", metrics).Value)
			assert.Equal(t, int64(1), *getMetricByName("PollCount", metrics).Delta) // Server sees delta=1

			collector.Collect()
			metrics = collector.GetMetrics()
			assert.Equal(t, tt.pollValue2, *getMetricByName("PollCount", metrics).Value)
			assert.Equal(t, int64(1), *getMetricByName("PollCount", metrics).Delta) // Server sees delta=1
		})
	}
}

func getNames(metrics []*models.Metrics) []string {
	names := []string{}
	for _, metric := range metrics {
		names = append(names, metric.ID)
	}
	return names
}

func getMetricByName(name string, metrics []*models.Metrics) *models.Metrics {
	for _, m := range metrics {
		if m.ID == name {
			return m
		}
	}
	return nil
}
