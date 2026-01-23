package agent

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			classifier := NewAgentErrorClassifier()
			sender := NewSender(baseURL, classifier)
			actualURL := sender.buildMetricURL(baseURL, tt.metricModel)
			assert.Equal(t, tt.expectedURL, actualURL)
		})
	}
}

func TestSendJSONWithRetries_Success(t *testing.T) {
	t.Run("succeeds on first attempt", func(t *testing.T) {
		var callCount int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&callCount, 1)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		classifier := NewAgentErrorClassifier()
		sender := NewSender(server.Listener.Addr().String(), classifier)

		metric := &models.Metrics{
			ID:    "test",
			MType: models.Gauge,
			Value: float64Ptr(42.0),
		}

		err := sender.SendJSONWithRetries(metric)

		assert.NoError(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&callCount), "should only make 1 attempt")
	})

	t.Run("succeeds on second attempt after 500 error", func(t *testing.T) {
		var callCount int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&callCount, 1)
			if count == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		classifier := NewAgentErrorClassifier()
		sender := NewSender(server.Listener.Addr().String(), classifier)

		metric := &models.Metrics{
			ID:    "test",
			MType: models.Gauge,
			Value: float64Ptr(42.0),
		}

		err := sender.SendJSONWithRetries(metric)

		assert.NoError(t, err)
		assert.Equal(t, int32(2), atomic.LoadInt32(&callCount), "should make 2 attempts")
	})

	t.Run("succeeds on third attempt after multiple 500 errors", func(t *testing.T) {
		var callCount int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&callCount, 1)
			if count < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		classifier := NewAgentErrorClassifier()
		sender := NewSender(server.Listener.Addr().String(), classifier)

		metric := &models.Metrics{
			ID:    "test",
			MType: models.Gauge,
			Value: float64Ptr(42.0),
		}

		err := sender.SendJSONWithRetries(metric)

		assert.NoError(t, err)
		assert.Equal(t, int32(3), atomic.LoadInt32(&callCount), "should make 3 attempts")
	})
}

func TestSendJSONWithRetries_Failure(t *testing.T) {
	t.Run("fails after max retries with 500 errors", func(t *testing.T) {
		var callCount int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&callCount, 1)
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		classifier := NewAgentErrorClassifier()
		sender := NewSender(server.Listener.Addr().String(), classifier)

		metric := &models.Metrics{
			ID:    "test",
			MType: models.Gauge,
			Value: float64Ptr(42.0),
		}

		err := sender.SendJSONWithRetries(metric)

		assert.Error(t, err)
		assert.Equal(t, int32(4), atomic.LoadInt32(&callCount), "should make 4 attempts")
		assert.Contains(t, err.Error(), "failed after 4 attempts")
	})

	t.Run("fails immediately on 400 error - no retry", func(t *testing.T) {
		var callCount int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&callCount, 1)
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		classifier := NewAgentErrorClassifier()
		sender := NewSender(server.Listener.Addr().String(), classifier)

		metric := &models.Metrics{
			ID:    "test",
			MType: models.Gauge,
			Value: float64Ptr(42.0),
		}

		err := sender.SendJSONWithRetries(metric)

		assert.Error(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&callCount), "should NOT retry on 4xx")
		assert.Contains(t, err.Error(), "non-retriable status: 400")
	})

	t.Run("fails immediately on 404 error - no retry", func(t *testing.T) {
		var callCount int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&callCount, 1)
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		classifier := NewAgentErrorClassifier()
		sender := NewSender(server.Listener.Addr().String(), classifier)

		metric := &models.Metrics{
			ID:    "test",
			MType: models.Gauge,
			Value: float64Ptr(42.0),
		}

		err := sender.SendJSONWithRetries(metric)

		assert.Error(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&callCount), "should NOT retry on 4xx")
		assert.Contains(t, err.Error(), "non-retriable status: 404")
	})
}

func TestSendJSONWithRetries_ConnectionErrors(t *testing.T) {
	t.Run("fails after max retries with connection errors", func(t *testing.T) {
		classifier := NewAgentErrorClassifier()
		sender := NewSender("localhost:9999", classifier)

		metric := &models.Metrics{
			ID:    "test",
			MType: models.Gauge,
			Value: float64Ptr(42.0),
		}

		err := sender.SendJSONWithRetries(metric)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed after 4 attempts")
	})
}

func TestSendJSONWithRetries_RetryDelays(t *testing.T) {
	t.Run("verifies retry delays are correct", func(t *testing.T) {
		timestamps := []time.Time{}
		var callCount int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timestamps = append(timestamps, time.Now())
			atomic.AddInt32(&callCount, 1)
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		classifier := NewAgentErrorClassifier()
		sender := NewSender(server.Listener.Addr().String(), classifier)

		metric := &models.Metrics{
			ID:    "test",
			MType: models.Gauge,
			Value: float64Ptr(42.0),
		}

		sender.SendJSONWithRetries(metric)

		require.Equal(t, 4, len(timestamps), "should have 4 attempts")

		tolerance := 200 * time.Millisecond

		delay1 := timestamps[1].Sub(timestamps[0])
		assert.InDelta(t, float64(1*time.Second), float64(delay1), float64(tolerance),
			"first retry should wait ~1s")

		delay2 := timestamps[2].Sub(timestamps[1])
		assert.InDelta(t, float64(3*time.Second), float64(delay2), float64(tolerance),
			"second retry should wait ~3s")

		delay3 := timestamps[3].Sub(timestamps[2])
		assert.InDelta(t, float64(5*time.Second), float64(delay3), float64(tolerance),
			"third retry should wait ~5s")
	})
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
