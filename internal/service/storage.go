// Package service содержит бизнес-логику и интерфейсы хранения метрик.
package service

import (
	"errors"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
)

// MetricStorage определяет интерфейс хранилища метрик.
type MetricStorage interface {
	// UpdateMetric обновляет или создаёт метрику. Возвращает true при успехе.
	UpdateMetric(metric *models.Metrics) bool
	// GetMetric возвращает метрику по имени или ошибку, если не найдена.
	GetMetric(metricName string) (*models.Metrics, error)
	// GetAllMetrics возвращает все сохранённые метрики.
	GetAllMetrics() []*models.Metrics
	// UpdateMetrics выполняет пакетное обновление метрик.
	UpdateMetrics(metrics []*models.Metrics) error
}

// ErrMetricNotFound возвращается, когда запрашиваемая метрика не найдена в хранилище.
var (
	ErrMetricNotFound = errors.New("metric not found")
)
