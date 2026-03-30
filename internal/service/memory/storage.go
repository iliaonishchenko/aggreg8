// Package memory реализует потокобезопасное хранилище метрик в оперативной памяти.
package memory

import (
	"errors"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/iliaonishchenko/aggreg8/internal/service"
	"sync"
)

// MemStorage — потокобезопасное in-memory хранилище метрик.
type MemStorage struct {
	mu      sync.RWMutex
	metrics map[string]*models.Metrics
}

// NewMemStorage создаёт новое пустое хранилище метрик в памяти.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]*models.Metrics),
	}
}

// UpdateMetric обновляет или создаёт метрику. Для Counter значения суммируются, для Gauge — перезаписываются.
func (ms *MemStorage) UpdateMetric(metric *models.Metrics) bool {
	ms.mu.Lock()
	defer ms.mu.Unlock()

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
		newDelta := *prev.Delta + *metric.Delta
		prev.Delta = &newDelta
		return true
	default:
		return false
	}
}

// GetMetric возвращает метрику по имени или ErrMetricNotFound, если не найдена.
func (ms *MemStorage) GetMetric(metricName string) (*models.Metrics, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	metric, exists := ms.metrics[metricName]
	if !exists {
		return nil, service.ErrMetricNotFound
	}

	return metric, nil
}

// GetAllMetrics возвращает срез всех сохранённых метрик.
func (ms *MemStorage) GetAllMetrics() []*models.Metrics {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	results := make([]*models.Metrics, 0, len(ms.metrics))
	for _, v := range ms.metrics {
		results = append(results, v)
	}

	return results
}

// UpdateMetrics выполняет пакетное обновление метрик.
func (ms *MemStorage) UpdateMetrics(metrics []*models.Metrics) error {
	for _, metric := range metrics {
		if ok := ms.UpdateMetric(metric); !ok {
			return errors.New("failed to update metric")
		}
	}
	return nil
}
