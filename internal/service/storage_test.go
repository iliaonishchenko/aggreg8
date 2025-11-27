package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdateGauge(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		value       float64
		secondValue float64
	}{
		{
			name:        "successfully update gauge metric",
			metricName:  "temperature",
			value:       23.5,
			secondValue: 47.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()
			err := storage.UpdateGauge(tt.metricName, tt.value)
			assert.NoError(t, err)
			value, exists := storage.GetGauge(tt.metricName)
			assert.True(t, exists)
			assert.Equal(t, tt.value, value)

			err = storage.UpdateGauge(tt.metricName, tt.secondValue)
			assert.NoError(t, err)
			value, exists = storage.GetGauge(tt.metricName)
			assert.True(t, exists)
			assert.Equal(t, tt.secondValue, value)
		})
	}
}

func TestUpdateCounter(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		value       int64
		secondValue int64
	}{
		{
			name:        "successfully update counter metric",
			metricName:  "requests",
			value:       10,
			secondValue: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()
			err := storage.UpdateCounter(tt.metricName, tt.value)
			assert.NoError(t, err)
			value, exists := storage.GetCounter(tt.metricName)
			assert.True(t, exists)
			assert.Equal(t, tt.value, value)

			err = storage.UpdateCounter(tt.metricName, tt.secondValue)
			assert.NoError(t, err)
			value, exists = storage.GetCounter(tt.metricName)
			assert.True(t, exists)
			assert.Equal(t, tt.value+tt.secondValue, value)
		})
	}
}
