package agent

import (
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"math/rand/v2"
	"runtime"
)

type Collector struct {
	pollCount int64
	Metrics   []*models.Metrics
}

func NewCollector() *Collector {
	return &Collector{}
}

func (c *Collector) Collect() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.Metrics = []*models.Metrics{
		{ID: "Alloc", MType: models.Gauge, Value: float64Ptr(float64(m.Alloc))},
		{ID: "BuckHashSys", MType: models.Gauge, Value: float64Ptr(float64(m.BuckHashSys))},
		{ID: "Frees", MType: models.Gauge, Value: float64Ptr(float64(m.Frees))},
		{ID: "GCCPUFraction", MType: models.Gauge, Value: float64Ptr(m.GCCPUFraction)},
		{ID: "GCSys", MType: models.Gauge, Value: float64Ptr(float64(m.GCSys))},
		{ID: "HeapAlloc", MType: models.Gauge, Value: float64Ptr(float64(m.HeapAlloc))},
		{ID: "HeapIdle", MType: models.Gauge, Value: float64Ptr(float64(m.HeapIdle))},
		{ID: "HeapInuse", MType: models.Gauge, Value: float64Ptr(float64(m.HeapInuse))},
		{ID: "HeapObjects", MType: models.Gauge, Value: float64Ptr(float64(m.HeapObjects))},
		{ID: "HeapReleased", MType: models.Gauge, Value: float64Ptr(float64(m.HeapReleased))},
		{ID: "HeapSys", MType: models.Gauge, Value: float64Ptr(float64(m.HeapSys))},
		{ID: "LastGC", MType: models.Gauge, Value: float64Ptr(float64(m.LastGC))},
		{ID: "Lookups", MType: models.Gauge, Value: float64Ptr(float64(m.Lookups))},
		{ID: "MCacheInuse", MType: models.Gauge, Value: float64Ptr(float64(m.MCacheInuse))},
		{ID: "MCacheSys", MType: models.Gauge, Value: float64Ptr(float64(m.MCacheSys))},
		{ID: "MSpanInuse", MType: models.Gauge, Value: float64Ptr(float64(m.MSpanInuse))},
		{ID: "MSpanSys", MType: models.Gauge, Value: float64Ptr(float64(m.MSpanSys))},
		{ID: "Mallocs", MType: models.Gauge, Value: float64Ptr(float64(m.Mallocs))},
		{ID: "NextGC", MType: models.Gauge, Value: float64Ptr(float64(m.NextGC))},
		{ID: "NumForcedGC", MType: models.Gauge, Value: float64Ptr(float64(m.NumForcedGC))},
		{ID: "NumGC", MType: models.Gauge, Value: float64Ptr(float64(m.NumGC))},
		{ID: "OtherSys", MType: models.Gauge, Value: float64Ptr(float64(m.OtherSys))},
		{ID: "PauseTotalNs", MType: models.Gauge, Value: float64Ptr(float64(m.PauseTotalNs))},
		{ID: "StackInuse", MType: models.Gauge, Value: float64Ptr(float64(m.StackInuse))},
		{ID: "StackSys", MType: models.Gauge, Value: float64Ptr(float64(m.StackSys))},
		{ID: "Sys", MType: models.Gauge, Value: float64Ptr(float64(m.Sys))},
		{ID: "TotalAlloc", MType: models.Gauge, Value: float64Ptr(float64(m.TotalAlloc))},
		{ID: "PollCount", MType: models.Counter, Delta: int64Ptr(c.pollCount)},
		{ID: "RandomValue", MType: models.Gauge, Value: float64Ptr(rand.Float64())},
	}
	c.pollCount++
}

func (c *Collector) GetMetrics() []*models.Metrics {
	result := make([]*models.Metrics, len(c.Metrics))
	copy(result, c.Metrics)
	return result
}

func float64Ptr(x float64) *float64 {
	return &x
}

func int64Ptr(x int64) *int64 {
	return &x
}
