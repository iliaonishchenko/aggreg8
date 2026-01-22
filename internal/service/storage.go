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

var (
	ErrMetricNotFound = errors.New("metric not found")
)
