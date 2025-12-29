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
	memStorage service.MetricStorage
}

func NewUpdateHandler(memStorage service.MetricStorage) *UpdateHandler {
	return &UpdateHandler{
		memStorage: memStorage,
	}
}

func (uh UpdateHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {

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
		ok := uh.memStorage.UpdateMetric(&metric)
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
		ok := uh.memStorage.UpdateMetric(&metric)
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

func (uh UpdateHandler) HandleUpdateJSON(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("decoding request")

	var metrics models.Metrics
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&metrics); err != nil {
		logger.Log.Error("cannot decode request JSON body", logger.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if metrics.ID == "" || metrics.MType == "" {
		logger.Log.Error("missing metric ID or type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if metrics.MType == models.Gauge && metrics.Value == nil {
		logger.Log.Error("missing gauge value")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if metrics.MType == models.Counter && metrics.Delta == nil {
		logger.Log.Error("missing counter delta")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if metrics.MType != models.Counter && metrics.MType != models.Gauge {
		logger.Log.Error("invalid metric type: " + metrics.MType)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if ok := uh.memStorage.UpdateMetric(&metrics); !ok {
		logger.Log.Error("failed to update metric")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
