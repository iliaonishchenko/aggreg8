package memory

import (
	"fmt"
	"testing"

	models "github.com/iliaonishchenko/aggreg8/internal/model"
)

func BenchmarkUpdateMetric_Gauge(b *testing.B) {
	storage := NewMemStorage()
	value := 42.5
	metric := &models.Metrics{ID: "temperature", MType: models.Gauge, Value: &value}

	for b.Loop() {
		storage.UpdateMetric(metric)
	}
}

func BenchmarkUpdateMetric_Counter(b *testing.B) {
	storage := NewMemStorage()
	delta := int64(1)
	metric := &models.Metrics{ID: "requests", MType: models.Counter, Delta: &delta}

	for b.Loop() {
		storage.UpdateMetric(metric)
	}
}

func BenchmarkUpdateMetrics_Batch(b *testing.B) {
	for _, size := range []int{1, 10, 100, 1000} {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			storage := NewMemStorage()
			metrics := make([]*models.Metrics, size)
			for i := range metrics {
				v := float64(i)
				metrics[i] = &models.Metrics{
					ID:    fmt.Sprintf("metric_%d", i),
					MType: models.Gauge,
					Value: &v,
				}
			}

			for b.Loop() {
				storage.UpdateMetrics(metrics)
			}
		})
	}
}

func BenchmarkGetAllMetrics(b *testing.B) {
	for _, size := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			storage := NewMemStorage()
			for i := 0; i < size; i++ {
				v := float64(i)
				storage.UpdateMetric(&models.Metrics{
					ID:    fmt.Sprintf("metric_%d", i),
					MType: models.Gauge,
					Value: &v,
				})
			}

			for b.Loop() {
				storage.GetAllMetrics()
			}
		})
	}
}
