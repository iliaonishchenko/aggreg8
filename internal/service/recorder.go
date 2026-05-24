package service

import (
	"errors"
	"time"

	"github.com/iliaonishchenko/aggreg8/internal/audit"
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
)

// ErrUpdateMetricRejected возвращается, когда хранилище отклонило одиночное обновление.
var ErrUpdateMetricRejected = errors.New("storage rejected metric update")

// Recorder инкапсулирует общую логику для всех транспортов (HTTP/gRPC):
// сохранение метрик в хранилище и оповещение наблюдателей аудита.
// Новые шаги обработки (метрики, трассировка, дополнительные наблюдатели) добавляются здесь,
// а не в каждом обработчике отдельно.
type Recorder struct {
	storage  MetricStorage
	notifier audit.Notifier
}

// NewRecorder создаёт Recorder поверх хранилища и нотификатора аудита.
func NewRecorder(storage MetricStorage, notifier audit.Notifier) *Recorder {
	return &Recorder{storage: storage, notifier: notifier}
}

// Record сохраняет батч метрик и оповещает наблюдателей аудита.
func (r *Recorder) Record(metrics []*models.Metrics, clientIP string) error {
	if len(metrics) == 0 {
		return nil
	}
	if err := r.storage.UpdateMetrics(metrics); err != nil {
		return err
	}
	r.notify(metrics, clientIP)
	return nil
}

// RecordOne сохраняет одиночную метрику и оповещает наблюдателей аудита.
func (r *Recorder) RecordOne(metric *models.Metrics, clientIP string) error {
	if metric == nil {
		return nil
	}
	if !r.storage.UpdateMetric(metric) {
		return ErrUpdateMetricRejected
	}
	r.notify([]*models.Metrics{metric}, clientIP)
	return nil
}

func (r *Recorder) notify(metrics []*models.Metrics, clientIP string) {
	if r.notifier == nil {
		return
	}
	names := make([]string, 0, len(metrics))
	for _, m := range metrics {
		names = append(names, m.ID)
	}
	event := audit.AuditEvent{
		TS:        time.Now().Unix(),
		Metrics:   names,
		IPAddress: clientIP,
	}
	for _, err := range r.notifier.NotifyAll(event) {
		logger.Log.Error("failed to notify audit", logger.Err(err))
	}
}
