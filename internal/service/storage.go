package service

type MetricStorage interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
}

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (ms *MemStorage) UpdateGauge(name string, value float64) error {
	ms.gauges[name] = value
	return nil
}
func (ms *MemStorage) UpdateCounter(name string, value int64) error {
	ms.counters[name] += value
	return nil
}
func (ms *MemStorage) GetGauge(name string) (float64, bool) {
	value, exists := ms.gauges[name]
	return value, exists
}

func (ms *MemStorage) GetCounter(name string) (int64, bool) {
	value, exists := ms.counters[name]
	return value, exists
}
