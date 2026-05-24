package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/iliaonishchenko/aggreg8/internal/service"
)

// UpdateHandler обрабатывает HTTP-запросы на создание и обновление метрик.
type UpdateHandler struct {
	recorder *service.Recorder
}

// NewUpdateHandler создаёт новый UpdateHandler с общим Recorder.
func NewUpdateHandler(recorder *service.Recorder) *UpdateHandler {
	return &UpdateHandler{recorder: recorder}
}

// HandleUpdate обрабатывает обновление метрики через URL-параметры: POST /update/{type}/{name}/{value}.
func (h UpdateHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {

	metricType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	if name == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var metric models.Metrics
	switch metricType {
	case models.Gauge:
		parsedValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		metric = models.Metrics{ID: name, MType: models.Gauge, Value: &parsedValue}
	case models.Counter:
		parsedValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		delta := parsedValue
		metric = models.Metrics{ID: name, MType: models.Counter, Delta: &delta}
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.recorder.RecordOne(&metric, r.RemoteAddr); err != nil {
		if errors.Is(err, service.ErrUpdateMetricRejected) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logger.Log.Error("failed to record metric", logger.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// HandleUpdateJSON обрабатывает обновление одной метрики через JSON-тело запроса: POST /update.
func (h UpdateHandler) HandleUpdateJSON(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("decoding request")

	var metrics models.Metrics
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metrics); err != nil {
		logger.Log.Error("cannot decode request JSON body", logger.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !h.isValidMetric(&metrics) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.recorder.RecordOne(&metrics, r.RemoteAddr); err != nil {
		logger.Log.Error("failed to update metric", logger.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleBatchUpdateJSON обрабатывает пакетное обновление метрик через JSON-массив: POST /updates.
func (h UpdateHandler) HandleBatchUpdateJSON(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("decoding batch request")

	var metrics []*models.Metrics
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metrics); err != nil {
		logger.Log.Error("cannot decode request JSON body", logger.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, metric := range metrics {
		if !h.isValidMetric(metric) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if err := h.recorder.Record(metrics, r.RemoteAddr); err != nil {
		logger.Log.Error("failed to update metrics batch", logger.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h UpdateHandler) isValidMetric(metric *models.Metrics) bool {
	if metric.ID == "" || metric.MType == "" {
		logger.Log.Error("missing metric ID or type")
		return false
	}

	if metric.MType == models.Gauge && metric.Value == nil {
		logger.Log.Error("missing gauge value")
		return false
	}

	if metric.MType == models.Counter && metric.Delta == nil {
		logger.Log.Error("missing counter delta")
		return false
	}

	if metric.MType != models.Counter && metric.MType != models.Gauge {
		logger.Log.Error("invalid metric type: " + metric.MType)
		return false
	}
	return true
}
