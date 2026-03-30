// Package pg реализует хранилище метрик на основе PostgreSQL.
package pg

import (
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
)

// MetricsRepository определяет интерфейс доступа к метрикам в базе данных.
type MetricsRepository interface {
	Ping() error
	Update(metric *models.Metrics) error
	Get(metricName string) (*models.Metrics, error)
	GetAll() ([]*models.Metrics, error)
	BatchUpdate(metrics []*models.Metrics) error
}

// PostgresStorage реализует MetricStorage поверх PostgreSQL через MetricsRepository.
type PostgresStorage struct {
	repo MetricsRepository
}

// NewPostgresStorage создаёт новое хранилище с указанным репозиторием.
func NewPostgresStorage(repo MetricsRepository) *PostgresStorage {
	return &PostgresStorage{repo: repo}
}

// UpdateMetric сохраняет метрику в PostgreSQL. Возвращает true при успехе.
func (s *PostgresStorage) UpdateMetric(metric *models.Metrics) (updated bool) {
	err := s.repo.Update(metric)
	if err != nil {
		logger.Log.Error("failed to update metric in database", logger.Err(err))
		return false
	}
	return true
}

// GetMetric получает метрику по имени из PostgreSQL.
func (s *PostgresStorage) GetMetric(metricName string) (*models.Metrics, error) {
	return s.repo.Get(metricName)
}

// GetAllMetrics возвращает все метрики из PostgreSQL.
func (s *PostgresStorage) GetAllMetrics() []*models.Metrics {
	metrics, err := s.repo.GetAll()
	if err != nil {
		logger.Log.Error("failed to get all metrics in database", logger.Err(err))
		return []*models.Metrics{}
	}
	return metrics
}

// UpdateMetrics выполняет пакетное обновление метрик в PostgreSQL.
func (s *PostgresStorage) UpdateMetrics(metrics []*models.Metrics) error {
	return s.repo.BatchUpdate(metrics)
}
