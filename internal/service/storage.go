package service

import (
	"errors"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
)

type MetricStorage interface {
	UpdateMetric(metric *models.Metrics) bool
	GetMetric(metricName string) (*models.Metrics, error)
	GetAllMetrics() []*models.Metrics
}

type MemStorage struct {
	metrics map[string]*models.Metrics
}

var ErrMetricNotFound = errors.New("metric not found")

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]*models.Metrics),
	}
}

func (ms *MemStorage) UpdateMetric(metric *models.Metrics) bool {
	switch metric.MType {
	case models.Gauge:
		prev, exists := ms.metrics[metric.ID]
		if !exists {
			ms.metrics[metric.ID] = metric
			return true
		}

		prev.Value = metric.Value
		return true
	case models.Counter:
		prev, exists := ms.metrics[metric.ID]
		if !exists {
			ms.metrics[metric.ID] = metric
			return true
		}
		newValue := *prev.Value + *metric.Value
		prev.Value = &newValue
		return true
	default:
		return false
	}
}

func (ms *MemStorage) GetMetric(metricName string) (*models.Metrics, error) {
	metric, exists := ms.metrics[metricName]
	if !exists {
		return nil, ErrMetricNotFound
	}
	return metric, nil
}

func (ms *MemStorage) GetAllMetrics() []*models.Metrics {
	results := make([]*models.Metrics, 0, len(ms.metrics))
	for _, v := range ms.metrics {
		results = append(results, v)
	}

	return results
}
