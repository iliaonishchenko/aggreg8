package file

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveToFile(t *testing.T) {
	t.Run("successfully save metrics to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "metrics.json")

		storage := NewFileStorage()
		metrics := []*models.Metrics{
			{ID: "temperature", MType: models.Gauge, Value: floatPtr(23.5)},
			{ID: "requests", MType: models.Counter, Delta: intPtr(42)},
		}

		err := storage.SaveToFile(filename, metrics)
		require.NoError(t, err)

		_, err = os.Stat(filename)
		assert.NoError(t, err)

		data, err := os.ReadFile(filename)
		require.NoError(t, err)

		var savedMetrics []*models.Metrics
		err = json.Unmarshal(data, &savedMetrics)
		require.NoError(t, err)

		assert.Len(t, savedMetrics, 2)
		assert.Equal(t, "temperature", savedMetrics[0].ID)
		assert.Equal(t, 23.5, *savedMetrics[0].Value)
		assert.Equal(t, "requests", savedMetrics[1].ID)
		assert.Equal(t, int64(42), *savedMetrics[1].Delta)
	})

	t.Run("save empty metrics array", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "empty.json")

		storage := NewFileStorage()
		metrics := []*models.Metrics{}

		err := storage.SaveToFile(filename, metrics)
		require.NoError(t, err)

		data, err := os.ReadFile(filename)
		require.NoError(t, err)

		var savedMetrics []*models.Metrics
		err = json.Unmarshal(data, &savedMetrics)
		require.NoError(t, err)

		assert.Len(t, savedMetrics, 0)
	})

	t.Run("overwrite existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "metrics.json")

		storage := NewFileStorage()

		metrics1 := []*models.Metrics{
			{ID: "temp1", MType: models.Gauge, Value: floatPtr(10.0)},
		}
		err := storage.SaveToFile(filename, metrics1)
		require.NoError(t, err)

		metrics2 := []*models.Metrics{
			{ID: "temp2", MType: models.Gauge, Value: floatPtr(20.0)},
			{ID: "temp3", MType: models.Gauge, Value: floatPtr(30.0)},
		}
		err = storage.SaveToFile(filename, metrics2)
		require.NoError(t, err)

		data, err := os.ReadFile(filename)
		require.NoError(t, err)

		var savedMetrics []*models.Metrics
		err = json.Unmarshal(data, &savedMetrics)
		require.NoError(t, err)

		assert.Len(t, savedMetrics, 2)
		assert.Equal(t, "temp2", savedMetrics[0].ID)
		assert.Equal(t, "temp3", savedMetrics[1].ID)
	})
}

func TestReadFromFile(t *testing.T) {
	t.Run("successfully read metrics from file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "metrics.json")

		metrics := []*models.Metrics{
			{ID: "temperature", MType: models.Gauge, Value: floatPtr(23.5)},
			{ID: "requests", MType: models.Counter, Delta: intPtr(42)},
		}
		data, err := json.Marshal(metrics)
		require.NoError(t, err)

		err = os.WriteFile(filename, data, 0644)
		require.NoError(t, err)

		storage := NewFileStorage()
		loadedMetrics, err := storage.ReadFromFile(filename)
		require.NoError(t, err)

		assert.Len(t, loadedMetrics, 2)
		assert.Equal(t, "temperature", loadedMetrics[0].ID)
		assert.Equal(t, 23.5, *loadedMetrics[0].Value)
		assert.Equal(t, "requests", loadedMetrics[1].ID)
		assert.Equal(t, int64(42), *loadedMetrics[1].Delta)
	})

	t.Run("file does not exist returns empty array", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "nonexistent.json")

		storage := NewFileStorage()
		metrics, err := storage.ReadFromFile(filename)

		require.NoError(t, err)
		assert.Empty(t, metrics)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "invalid.json")

		err := os.WriteFile(filename, []byte("invalid json"), 0644)
		require.NoError(t, err)

		storage := NewFileStorage()
		_, err = storage.ReadFromFile(filename)

		assert.Error(t, err)
	})

	t.Run("empty file returns empty array", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "empty.json")

		err := os.WriteFile(filename, []byte("[]"), 0644)
		require.NoError(t, err)

		storage := NewFileStorage()
		metrics, err := storage.ReadFromFile(filename)

		require.NoError(t, err)
		assert.Empty(t, metrics)
	})
}

func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int64) *int64 {
	return &i
}
