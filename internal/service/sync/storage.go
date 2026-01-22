package sync

import (
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/iliaonishchenko/aggreg8/internal/service"
)

type MetricSaver interface {
	SaveToFile(filename string, metrics []*models.Metrics) error
}

type SyncStorage struct {
	storage  service.MetricStorage
	saver    MetricSaver
	filename string
}

func NewSyncStorage(storage service.MetricStorage, saver MetricSaver, filename string) *SyncStorage {
	return &SyncStorage{
		storage:  storage,
		saver:    saver,
		filename: filename,
	}
}

func (ss *SyncStorage) UpdateMetric(metric *models.Metrics) bool {
	result := ss.storage.UpdateMetric(metric)

	if result {
		metrics := ss.storage.GetAllMetrics()
		_ = ss.saver.SaveToFile(ss.filename, metrics)
	}

	return result
}

func (ss *SyncStorage) GetMetric(metricName string) (*models.Metrics, error) {
	return ss.storage.GetMetric(metricName)
}

func (ss *SyncStorage) GetAllMetrics() []*models.Metrics {
	return ss.storage.GetAllMetrics()
}

func (ss *SyncStorage) UpdateMetrics(metrics []*models.Metrics) error {
	err := ss.storage.UpdateMetrics(metrics)
	if err != nil {
		return err
	}
	latestMetrics := ss.storage.GetAllMetrics()
	return ss.saver.SaveToFile(ss.filename, latestMetrics)
}
