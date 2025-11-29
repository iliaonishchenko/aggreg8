package agent

import (
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCollect(t *testing.T) {
	t.Run("collects all expected metrics", func(t *testing.T) {
		collector := NewCollector()
		collector.Collect()

		metrics := collector.GetMetrics()
		assert.Contains(t, getNames(metrics), "Alloc")
		assert.Contains(t, getNames(metrics), "PollCount")
		assert.Contains(t, getNames(metrics), "RandomValue")
	})

	t.Run("PollCount delta accumulates", func(t *testing.T) {
		collector := NewCollector()

		collector.Collect()
		metrics := collector.GetMetrics()
		pollCount := getMetricByName("PollCount", metrics)
		assert.Equal(t, models.Counter, pollCount.MType)
		assert.Equal(t, int64(0), *pollCount.Delta)

		collector.Collect()
		metrics = collector.GetMetrics()
		pollCount = getMetricByName("PollCount", metrics)
		assert.Equal(t, int64(1), *pollCount.Delta)
	})
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
