package sync

import (
	"github.com/iliaonishchenko/aggreg8/internal/service/memory"
	"testing"

	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMetricSaver struct {
	saveToFileCalled bool
	saveToFileError  error
	savedMetrics     []*models.Metrics
	savedFilename    string
}

func (m *mockMetricSaver) SaveToFile(filename string, metrics []*models.Metrics) error {
	m.saveToFileCalled = true
	m.savedFilename = filename
	m.savedMetrics = metrics
	return m.saveToFileError
}

func TestUpdateMetric(t *testing.T) {
	t.Run("successfully update metric and save to file", func(t *testing.T) {
		baseStorage := memory.NewMemStorage()
		mockSaver := &mockMetricSaver{}
		syncStorage := NewSyncStorage(baseStorage, mockSaver, "test.json")

		metric := &models.Metrics{
			ID:    "temperature",
			MType: models.Gauge,
			Value: floatPtr(23.5),
		}

		result := syncStorage.UpdateMetric(metric)

		assert.True(t, result)
		assert.True(t, mockSaver.saveToFileCalled, "SaveToFile should be called")
		assert.Equal(t, "test.json", mockSaver.savedFilename)
		assert.Len(t, mockSaver.savedMetrics, 1)
		assert.Equal(t, "temperature", mockSaver.savedMetrics[0].ID)
		assert.Equal(t, 23.5, *mockSaver.savedMetrics[0].Value)
	})

	t.Run("update multiple metrics saves all", func(t *testing.T) {
		baseStorage := memory.NewMemStorage()
		mockSaver := &mockMetricSaver{}
		syncStorage := NewSyncStorage(baseStorage, mockSaver, "test.json")

		metric1 := &models.Metrics{ID: "temp1", MType: models.Gauge, Value: floatPtr(10.0)}
		metric2 := &models.Metrics{ID: "temp2", MType: models.Gauge, Value: floatPtr(20.0)}

		syncStorage.UpdateMetric(metric1)
		mockSaver.saveToFileCalled = false

		syncStorage.UpdateMetric(metric2)

		assert.True(t, mockSaver.saveToFileCalled)
		assert.Len(t, mockSaver.savedMetrics, 2, "Should save all metrics")
	})
}

func TestGetMetric(t *testing.T) {
	t.Run("delegates to base storage", func(t *testing.T) {
		baseStorage := memory.NewMemStorage()
		mockSaver := &mockMetricSaver{}
		syncStorage := NewSyncStorage(baseStorage, mockSaver, "test.json")

		metric := &models.Metrics{ID: "temp", MType: models.Gauge, Value: floatPtr(10.0)}
		baseStorage.UpdateMetric(metric)

		result, err := syncStorage.GetMetric("temp")

		require.NoError(t, err)
		assert.Equal(t, "temp", result.ID)
		assert.Equal(t, 10.0, *result.Value)
	})
}

func TestGetAllMetrics(t *testing.T) {
	t.Run("delegates to base storage", func(t *testing.T) {
		baseStorage := memory.NewMemStorage()
		mockSaver := &mockMetricSaver{}
		syncStorage := NewSyncStorage(baseStorage, mockSaver, "test.json")

		metric1 := &models.Metrics{ID: "temp1", MType: models.Gauge, Value: floatPtr(10.0)}
		metric2 := &models.Metrics{ID: "temp2", MType: models.Gauge, Value: floatPtr(20.0)}
		syncStorage.UpdateMetric(metric1)
		syncStorage.UpdateMetric(metric2)

		results := syncStorage.GetAllMetrics()

		assert.Len(t, results, 2)
	})
}
func floatPtr(f float64) *float64 {
	return &f
}
