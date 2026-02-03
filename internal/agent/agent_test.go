package agent

import (
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

type MockCollector struct {
	mu              sync.Mutex
	collectCount    int
	systemCount     int
	metricsToReturn [][]*models.Metrics
}

func (m *MockCollector) CollectRuntime() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.collectCount++
}

func (m *MockCollector) CollectSystem() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.systemCount++
}

func (m *MockCollector) GetMetrics() [][]*models.Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := m.metricsToReturn
	m.metricsToReturn = nil
	return result
}

type MockSender struct {
	mu        sync.Mutex
	sendCount int
	calls     [][]*models.Metrics
	delay     time.Duration
}

func (m *MockSender) SendJSONWithRetries(metrics ...*models.Metrics) error {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendCount++
	m.calls = append(m.calls, metrics)
	return nil
}

func TestRun(t *testing.T) {
	t.Run("collects on poll interval", func(t *testing.T) {
		mockCollector := &MockCollector{}
		mockSender := &MockSender{}
		agent := NewAgent(mockCollector, mockSender, 1, 2, 2)

		stopCh := make(chan struct{})
		go func() {
			time.Sleep(3 * time.Second)
			close(stopCh)
		}()

		go func() {
			agent.Run()
		}()

		<-stopCh

		assert.GreaterOrEqual(t, mockCollector.collectCount, 2)
		assert.GreaterOrEqual(t, mockCollector.systemCount, 2)
	})

	t.Run("sends on report interval", func(t *testing.T) {
		metric := &models.Metrics{ID: "test", MType: models.Gauge, Value: float64Ptr(1.0)}
		mockCollector := &MockCollector{
			metricsToReturn: [][]*models.Metrics{{metric}},
		}
		mockSender := &MockSender{}
		agent := NewAgent(mockCollector, mockSender, 1, 2, 2)

		stopCh := make(chan struct{})
		go func() {
			time.Sleep(3 * time.Second)
			close(stopCh)
		}()

		go func() {
			agent.Run()
		}()

		<-stopCh

		assert.Equal(t, 1, mockSender.sendCount)
	})
}
