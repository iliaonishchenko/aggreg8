package pg

import (
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
)

type MetricsRepository interface {
	Ping() error
	Update(metric *models.Metrics) error
	Get(metricName string) (*models.Metrics, error)
	GetAll() ([]*models.Metrics, error)
	BatchUpdate(metrics []*models.Metrics) error
}

type PostgresStorage struct {
	repo MetricsRepository
}

func NewPostgresStorage(repo MetricsRepository) *PostgresStorage {
	return &PostgresStorage{repo: repo}
}

func (s *PostgresStorage) UpdateMetric(metric *models.Metrics) (updated bool) {
	err := s.repo.Update(metric)
	if err != nil {
		logger.Log.Error("failed to update metric in database", logger.Err(err))
		return false
	}
	return true
}

func (s *PostgresStorage) GetMetric(metricName string) (*models.Metrics, error) {
	return s.repo.Get(metricName)
}

func (s *PostgresStorage) GetAllMetrics() []*models.Metrics {
	metrics, err := s.repo.GetAll()
	if err != nil {
		logger.Log.Error("failed to get all metrics in database", logger.Err(err))
		return []*models.Metrics{}
	}
	return metrics
}

func (s *PostgresStorage) UpdateMetrics(metrics []*models.Metrics) error {
	return s.repo.BatchUpdate(metrics)
}
