package agent

import (
	"compress/gzip"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestBuildMetricURL(t *testing.T) {
	baseURL := "localhost:8080"

	tests := []struct {
		name        string
		metricModel *models.Metrics
		expectedURL string
		err         error
	}{
		{
			name: "build gauge metric",
			metricModel: &models.Metrics{
				ID:    "temperature",
				MType: models.Gauge,
				Value: float64Ptr(23.5),
			},
			expectedURL: "http://localhost:8080/update/gauge/temperature/23.5",
			err:         nil,
		},
		{
			name: "build counter metric",
			metricModel: &models.Metrics{
				ID:    "requests",
				MType: models.Counter,
				Delta: int64Ptr(42),
			},
			expectedURL: "http://localhost:8080/update/counter/requests/42",
			err:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := NewSender(baseURL)
			actualURL := sender.buildMetricURL(baseURL, tt.metricModel)
			assert.Equal(t, tt.expectedURL, actualURL)
		})
	}
}

func TestCompressData(t *testing.T) {
	original := []byte("This is some test data to be compressed.")
	compressedData, err := compressData(original)

	assert.NoError(t, err)
	assert.NotNil(t, compressedData)

	reader, err := gzip.NewReader(compressedData)
	assert.NoError(t, err)

	decompressed, err := io.ReadAll(reader)

	assert.NoError(t, err)
	assert.Equal(t, original, decompressed)
}

func TestEncodeJSON(t *testing.T) {
	metric := &models.Metrics{
		ID:    "cpu_usage",
		MType: models.Gauge,
		Value: float64Ptr(75.5),
	}

	jsonData, err := encodeJSON(metric)

	assert.NoError(t, err)
	assert.NotNil(t, jsonData)

	expectedJSON := `{"id":"cpu_usage","type":"gauge","value":75.5}`
	assert.JSONEq(t, expectedJSON, jsonData.String())
}
