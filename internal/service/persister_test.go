package service

import (
	"context"
	"sync"
	"testing"
	"time"

	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/stretchr/testify/assert"
)

type mockMetricReader struct {
	mu      sync.Mutex
	metrics []*models.Metrics
}

func (m *mockMetricReader) GetAllMetrics() []*models.Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.metrics
}

func (m *mockMetricReader) addMetric(metric *models.Metrics) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics = append(m.metrics, metric)
}

type mockFileSaver struct {
	mu            sync.Mutex
	saveCount     int
	savedMetrics  [][]*models.Metrics
	savedFilename string
	saveError     error
}

func (m *mockFileSaver) SaveToFile(filename string, metrics []*models.Metrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.saveCount++
	m.savedFilename = filename

	metricsCopy := make([]*models.Metrics, len(metrics))
	copy(metricsCopy, metrics)
	m.savedMetrics = append(m.savedMetrics, metricsCopy)

	return m.saveError
}

func (m *mockFileSaver) getSaveCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveCount
}

func (m *mockFileSaver) getLastSaved() []*models.Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.savedMetrics) == 0 {
		return nil
	}
	return m.savedMetrics[len(m.savedMetrics)-1]
}

func TestPersister_Start(t *testing.T) {
	t.Run("saves metrics periodically", func(t *testing.T) {
		reader := &mockMetricReader{
			metrics: []*models.Metrics{
				{ID: "temp", MType: models.Gauge, Value: floatPtr(23.5)},
			},
		}
		saver := &mockFileSaver{}
		interval := 100 * time.Millisecond
		persister := NewPersister(reader, saver, "test.json", interval)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go persister.Start(ctx)

		time.Sleep(250 * time.Millisecond)
		cancel()

		time.Sleep(50 * time.Millisecond)

		saveCount := saver.getSaveCount()
		assert.GreaterOrEqual(t, saveCount, 2, "Should save at least 2 times")
		assert.Equal(t, "test.json", saver.savedFilename)
	})

	t.Run("saves updated metrics", func(t *testing.T) {
		reader := &mockMetricReader{
			metrics: []*models.Metrics{
				{ID: "counter", MType: models.Counter, Delta: intPtr(1)},
			},
		}
		saver := &mockFileSaver{}
		interval := 100 * time.Millisecond
		persister := NewPersister(reader, saver, "test.json", interval)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go persister.Start(ctx)

		time.Sleep(150 * time.Millisecond)

		reader.addMetric(&models.Metrics{ID: "new", MType: models.Gauge, Value: floatPtr(10.0)})

		time.Sleep(150 * time.Millisecond)
		cancel()

		time.Sleep(50 * time.Millisecond)

		lastSaved := saver.getLastSaved()
		assert.GreaterOrEqual(t, len(lastSaved), 1, "Should save updated metrics")
	})

	t.Run("continues on save error", func(t *testing.T) {
		reader := &mockMetricReader{
			metrics: []*models.Metrics{
				{ID: "temp", MType: models.Gauge, Value: floatPtr(23.5)},
			},
		}
		saver := &mockFileSaver{
			saveError: assert.AnError,
		}
		interval := 100 * time.Millisecond
		persister := NewPersister(reader, saver, "test.json", interval)

		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		defer cancel()

		err := persister.Start(ctx)

		assert.Error(t, err)
		assert.GreaterOrEqual(t, saver.getSaveCount(), 2, "Should continue trying to save despite errors")
	})
}
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int64) *int64 {
	return &i
}
