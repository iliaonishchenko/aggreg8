package pg

import (
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/iliaonishchenko/aggreg8/internal/repository"
)

type PostgresStorage struct {
	repo repository.Repository
}

func NewPostgresStorage(repo repository.Repository) *PostgresStorage {
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
