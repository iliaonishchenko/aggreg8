package agent

import (
	"context"
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"time"
)

// CollectorService определяет интерфейс сбора метрик (runtime и system).
type CollectorService interface {
	CollectRuntime()
	CollectSystem()
	GetMetrics() [][]*models.Metrics
}

// SenderService определяет интерфейс отправки метрик на сервер с ретраями.
type SenderService interface {
	SendJSONWithRetries(metrics ...*models.Metrics) error
}

// Agent координирует периодический сбор и отправку метрик с ограничением параллелизма через rateLimit.
type Agent struct {
	collector      CollectorService
	sender         SenderService
	pollInterval   int
	reportInterval int
	rateLimit      int
}

// NewAgent создаёт нового агента с указанными интервалами сбора/отправки и лимитом воркеров.
func NewAgent(collector CollectorService, sender SenderService, pollInterval, reportInterval, rateLimit int) *Agent {
	return &Agent{
		collector:      collector,
		sender:         sender,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		rateLimit:      rateLimit,
	}
}

// Run запускает цикл сбора и отправки метрик. Блокирует выполнение до отмены контекста.
func (a *Agent) Run(ctx context.Context) {
	collectTicker := time.NewTicker(time.Duration(a.pollInterval) * time.Second)
	defer collectTicker.Stop()
	reportTicker := time.NewTicker(time.Duration(a.reportInterval) * time.Second)
	defer reportTicker.Stop()
	jobsCh := make(chan []*models.Metrics, a.rateLimit)

	for i := 0; i < a.rateLimit; i++ {
		go a.worker(jobsCh)
	}

	for {
		select {
		case <-collectTicker.C:
			go a.collector.CollectRuntime()
			go a.collector.CollectSystem()
		case <-reportTicker.C:
			go func() {
				metrics := a.collector.GetMetrics()
				for _, metricBatch := range metrics {
					jobsCh <- metricBatch
				}
			}()
		case <-ctx.Done():
			logger.Log.Info("Shutting down agent...")
			close(jobsCh)
			return
		}
	}
}

func (a *Agent) worker(jobs <-chan []*models.Metrics) {
	for metrics := range jobs {
		logger.Log.Info("Sending metrics %v", logger.Field("metrics", metrics[0]))
		err := a.sender.SendJSONWithRetries(metrics...)
		if err != nil {
			logger.Log.Error("error sending metrics: %v", logger.Err(err))
		}
	}
}
