package handler

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/iliaonishchenko/aggreg8/internal/service"
	"net/http"
	"strconv"
)

type UpdateHandler struct {
	storage service.MetricStorage
}

func NewUpdateHandler(memStorage service.MetricStorage) *UpdateHandler {
	return &UpdateHandler{
		storage: memStorage,
	}
}

func (h UpdateHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {

	metricType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	if name == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch metricType {
	case models.Gauge:
		parsedValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		metric := models.Metrics{ID: name, MType: models.Gauge, Value: &parsedValue}
		ok := h.storage.UpdateMetric(&metric)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case models.Counter:
		parsedValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		delta := parsedValue
		metric := models.Metrics{ID: name, MType: models.Counter, Delta: &delta}
		ok := h.storage.UpdateMetric(&metric)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

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

	if ok := h.storage.UpdateMetric(&metrics); !ok {
		logger.Log.Error("failed to update metric")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

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
	err := h.storage.UpdateMetrics(metrics)
	if err != nil {
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
