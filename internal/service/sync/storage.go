// Package sync реализует обёртку над хранилищем метрик с синхронной записью в файл при каждом обновлении.
package sync

import (
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/iliaonishchenko/aggreg8/internal/service"
)

// MetricSaver определяет интерфейс для сохранения метрик в файл.
type MetricSaver interface {
	SaveToFile(filename string, metrics []*models.Metrics) error
}

// SyncStorage — обёртка над MetricStorage, которая синхронно сохраняет метрики в файл после каждого обновления.
type SyncStorage struct {
	storage  service.MetricStorage
	saver    MetricSaver
	filename string
}

// NewSyncStorage создаёт новый SyncStorage с указанным хранилищем, сейвером и файлом.
func NewSyncStorage(storage service.MetricStorage, saver MetricSaver, filename string) *SyncStorage {
	return &SyncStorage{
		storage:  storage,
		saver:    saver,
		filename: filename,
	}
}

// UpdateMetric обновляет метрику и синхронно сохраняет все метрики в файл.
func (ss *SyncStorage) UpdateMetric(metric *models.Metrics) bool {
	result := ss.storage.UpdateMetric(metric)

	if result {
		metrics := ss.storage.GetAllMetrics()
		_ = ss.saver.SaveToFile(ss.filename, metrics)
	}

	return result
}

// GetMetric возвращает метрику по имени из базового хранилища.
func (ss *SyncStorage) GetMetric(metricName string) (*models.Metrics, error) {
	return ss.storage.GetMetric(metricName)
}

// GetAllMetrics возвращает все метрики из базового хранилища.
func (ss *SyncStorage) GetAllMetrics() []*models.Metrics {
	return ss.storage.GetAllMetrics()
}

// UpdateMetrics выполняет пакетное обновление и синхронно сохраняет результат в файл.
func (ss *SyncStorage) UpdateMetrics(metrics []*models.Metrics) error {
	err := ss.storage.UpdateMetrics(metrics)
	if err != nil {
		return err
	}
	latestMetrics := ss.storage.GetAllMetrics()
	return ss.saver.SaveToFile(ss.filename, latestMetrics)
}
