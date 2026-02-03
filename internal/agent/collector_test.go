package agent

import (
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCollectRuntime(t *testing.T) {
	t.Run("collects all expected runtime metrics", func(t *testing.T) {
		collector := NewCollector()
		collector.CollectRuntime()

		metrics := collector.GetMetrics()[0]
		assert.Contains(t, getNames(metrics), "Alloc")
		assert.Contains(t, getNames(metrics), "PollCount")
		assert.Contains(t, getNames(metrics), "RandomValue")
	})

	t.Run("PollCount delta accumulates", func(t *testing.T) {
		collector := NewCollector()

		collector.CollectRuntime()
		metrics := collector.GetMetrics()[0]
		pollCount := getMetricByName("PollCount", metrics)
		assert.Equal(t, models.Counter, pollCount.MType)
		assert.Equal(t, int64(0), *pollCount.Delta)

		collector.CollectRuntime()
		metrics = collector.GetMetrics()[0]
		pollCount = getMetricByName("PollCount", metrics)
		assert.Equal(t, int64(1), *pollCount.Delta)
	})
	t.Run("Runtime metrics accumulate", func(t *testing.T) {
		collector := NewCollector()

		collector.CollectRuntime()
		collector.CollectRuntime()
		metrics := collector.GetMetrics()
		assert.Equal(t, 2, len(metrics))
		assert.Contains(t, getNames(metrics[0]), "Alloc")
		assert.Contains(t, getNames(metrics[0]), "PollCount")
		assert.Contains(t, getNames(metrics[0]), "RandomValue")
		assert.Contains(t, getNames(metrics[1]), "Alloc")
		assert.Contains(t, getNames(metrics[1]), "PollCount")
		assert.Contains(t, getNames(metrics[1]), "RandomValue")
	})
}

func TestCollectSystem(t *testing.T) {
	t.Run("collects system metrics", func(t *testing.T) {
		collector := NewCollector()
		collector.CollectSystem()

		metrics := collector.GetMetrics()[0]
		assert.Contains(t, getNames(metrics), "TotalMemory")
		assert.Contains(t, getNames(metrics), "FreeMemory")
		assert.Contains(t, getNames(metrics), "CPUutilization1")
	})
	t.Run("System metrics accumulate", func(t *testing.T) {
		collector := NewCollector()

		collector.CollectSystem()
		collector.CollectSystem()
		metrics := collector.GetMetrics()
		assert.Equal(t, 2, len(metrics))
		assert.Contains(t, getNames(metrics[0]), "TotalMemory")
		assert.Contains(t, getNames(metrics[0]), "FreeMemory")
		assert.Contains(t, getNames(metrics[0]), "CPUutilization1")
		assert.Contains(t, getNames(metrics[1]), "TotalMemory")
		assert.Contains(t, getNames(metrics[1]), "FreeMemory")
		assert.Contains(t, getNames(metrics[1]), "CPUutilization1")
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
