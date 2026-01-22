package memory

import (
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdateMetric(t *testing.T) {
	t.Run("successfully update gauge metric", func(t *testing.T) {
		storage := NewMemStorage()
		metric := &models.Metrics{ID: "temperature", MType: models.Gauge, Value: floatPtr(23.5)}

		ok := storage.UpdateMetric(metric)
		assert.True(t, ok)

		storedMetric, err := storage.GetMetric("temperature")
		assert.NoError(t, err)
		assert.Equal(t, 23.5, *storedMetric.Value)

		metric2 := &models.Metrics{ID: "temperature", MType: models.Gauge, Value: floatPtr(99.9)}
		ok = storage.UpdateMetric(metric2)
		assert.True(t, ok)

		storedMetric, err = storage.GetMetric("temperature")
		assert.NoError(t, err)
		assert.Equal(t, 99.9, *storedMetric.Value)
	})

	t.Run("successfully update counter metric", func(t *testing.T) {
		storage := NewMemStorage()
		metric := &models.Metrics{ID: "requests", MType: models.Counter, Delta: intPtr(1)}

		ok := storage.UpdateMetric(metric)
		assert.True(t, ok)

		storedMetric, err := storage.GetMetric("requests")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), *storedMetric.Delta)

		ok = storage.UpdateMetric(metric)
		assert.True(t, ok)

		storedMetric, err = storage.GetMetric("requests")
		assert.NoError(t, err)
		assert.Equal(t, int64(2), *storedMetric.Delta)
	})

	t.Run("unknown metric type returns false", func(t *testing.T) {
		storage := NewMemStorage()
		metric := &models.Metrics{ID: "unknown", MType: "invalid"}

		ok := storage.UpdateMetric(metric)
		assert.False(t, ok)
	})
}

func TestGetAllMetrics(t *testing.T) {
	tests := []struct {
		name    string
		metrics []*models.Metrics
	}{
		{
			name: "successfully get all metrics",
			metrics: []*models.Metrics{
				{ID: "temperature", MType: models.Gauge, Value: floatPtr(23.5)},
				{ID: "requests", MType: models.Counter, Delta: intPtr(1)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()

			for _, metric := range tt.metrics {
				storage.UpdateMetric(metric)
			}

			allMetrics := storage.GetAllMetrics()

			assert.Equal(t, len(tt.metrics), len(allMetrics))
			for _, metric := range tt.metrics {
				found := false
				for _, storedMetric := range allMetrics {
					if storedMetric.ID == metric.ID && storedMetric.MType == metric.MType {
						found = true
						break
					}
				}
				assert.True(t, found, "metric not found: %v", metric)
			}
		})
	}
}

func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int64) *int64 {
	return &i
}
