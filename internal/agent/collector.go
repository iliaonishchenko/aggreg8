package agent

import (
	"fmt"
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	cpu2 "github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"math/rand/v2"
	"runtime"
	"sync"
)

type Collector struct {
	pollCount int64
	metrics   [][]*models.Metrics
	mu        sync.Mutex
}

func NewCollector() *Collector {
	return &Collector{}
}

func (c *Collector) CollectRuntime() {
	logger.Log.Info("Collecting runtime metrics")
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.mu.Lock()
	defer c.mu.Unlock()
	newMetrics := []*models.Metrics{
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
	c.metrics = append(c.metrics, newMetrics)
	c.pollCount++
}

func (c *Collector) CollectSystem() {
	logger.Log.Info("Collecting system metrics")
	v, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Error("Error collecting memory metrics", logger.Err(err))
		return
	}
	perCPUs, err := cpu2.Percent(0, true)
	if err != nil {
		logger.Log.Error("Error collecting CPU metrics", logger.Err(err))
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	newMetrics := []*models.Metrics{
		{ID: "TotalMemory", MType: models.Gauge, Value: float64Ptr(float64(v.Total))},
		{ID: "FreeMemory", MType: models.Gauge, Value: float64Ptr(float64(v.Free))},
	}
	for i, cpu := range perCPUs {
		newMetrics = append(newMetrics, &models.Metrics{
			ID:    fmt.Sprintf("CPUutilization%d", i+1),
			MType: models.Gauge,
			Value: float64Ptr(cpu),
		})
	}
	c.metrics = append(c.metrics, newMetrics)
}

func (c *Collector) GetMetrics() [][]*models.Metrics {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([][]*models.Metrics, len(c.metrics))
	copy(result, c.metrics)
	c.metrics = [][]*models.Metrics{}
	return result
}

func float64Ptr(x float64) *float64 {
	return &x
}

func int64Ptr(x int64) *int64 {
	return &x
}
