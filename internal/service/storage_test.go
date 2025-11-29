package service

import (
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdateMetric(t *testing.T) {
	tests := []struct {
		name              string
		metric            *models.Metrics
		updateResult      bool
		storedMetricValue float64
	}{
		{
			name:              "successfully update gauge metric",
			metric:            &models.Metrics{ID: "temperature", MType: models.Gauge, Value: floatPtr(23.5)},
			updateResult:      true,
			storedMetricValue: 23.5,
		},
		{
			name:              "successfully update counter metric",
			metric:            &models.Metrics{ID: "requests", MType: models.Counter, Delta: intPtr(1), Value: floatPtr(1)},
			updateResult:      true,
			storedMetricValue: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()

			actualUpdateResult := storage.UpdateMetric(tt.metric)

			assert.Equal(t, tt.updateResult, actualUpdateResult)

			storedMetric, err := storage.GetMetric(tt.metric.ID)

			assert.NoError(t, err)
			assert.Equal(t, storedMetric, tt.metric)

			actualUpdateResult2 := storage.UpdateMetric(tt.metric)

			assert.Equal(t, tt.updateResult, actualUpdateResult2)

			storedMetric2, err := storage.GetMetric(tt.metric.ID)

			assert.NoError(t, err)
			assert.Equal(t, tt.storedMetricValue, *storedMetric2.Value)
		})
	}
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
				{ID: "requests", MType: models.Counter, Delta: intPtr(1), Value: floatPtr(1)},
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
