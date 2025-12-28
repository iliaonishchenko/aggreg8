package service

import (
	"context"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"time"
)

type MetricReader interface {
	GetAllMetrics() []*models.Metrics
}

type FileSaver interface {
	SaveToFile(filename string, metrics []*models.Metrics) error
}

type Persister struct {
	reader   MetricReader
	saver    FileSaver
	filename string
	interval time.Duration
}

func NewPersister(reader MetricReader, saver FileSaver, filename string, interval time.Duration) *Persister {
	return &Persister{
		reader:   reader,
		saver:    saver,
		filename: filename,
		interval: interval,
	}
}

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
