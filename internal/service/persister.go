package service

import (
	"context"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"time"
)

// MetricReader определяет интерфейс для чтения всех метрик из хранилища.
type MetricReader interface {
	GetAllMetrics() []*models.Metrics
}

// FileSaver определяет интерфейс для сохранения метрик в файл.
type FileSaver interface {
	SaveToFile(filename string, metrics []*models.Metrics) error
}

// Persister периодически сохраняет метрики из хранилища в файл с заданным интервалом.
type Persister struct {
	reader   MetricReader
	saver    FileSaver
	filename string
	interval time.Duration
}

// NewPersister создаёт новый Persister с указанными параметрами.
func NewPersister(reader MetricReader, saver FileSaver, filename string, interval time.Duration) *Persister {
	return &Persister{
		reader:   reader,
		saver:    saver,
		filename: filename,
		interval: interval,
	}
}

// Start запускает цикл периодического сохранения метрик. Блокирует выполнение до отмены контекста.
func (p *Persister) Start(ctx context.Context) error {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics := p.reader.GetAllMetrics()
			_ = p.saver.SaveToFile(p.filename, metrics)
		case <-ctx.Done():
			metrics := p.reader.GetAllMetrics()
			return p.saver.SaveToFile(p.filename, metrics)
		}
	}
}
