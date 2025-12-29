package handler

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/iliaonishchenko/aggreg8/internal/service"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
)

type GetMetricHandler struct {
	storage service.MetricStorage
}

func NewGetMetricHandler(s service.MetricStorage) *GetMetricHandler {
	return &GetMetricHandler{
		storage: s,
	}
}

func (gmh *GetMetricHandler) HandleGetMetric(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	var resultValue string

	switch mType {
	case models.Gauge:
		metric, err := gmh.storage.GetMetric(name)
		if err != nil {
			http.Error(w, "unknown metric: "+name, http.StatusNotFound)
			return
		}
		if metric.Value == nil {
			http.Error(w, "invalid metric data", http.StatusInternalServerError)
			return
		}
		resultValue = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	case models.Counter:
		metric, err := gmh.storage.GetMetric(name)
		if err != nil {
			http.Error(w, "unknown metric: "+name, http.StatusNotFound)
			return
		}
		if metric.Delta == nil {
			http.Error(w, "invalid metric data", http.StatusInternalServerError)
			return
		}
		resultValue = strconv.FormatInt(int64(*metric.Delta), 10)
	default:
		http.Error(w, "unknown metric type: "+mType, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, resultValue)
}

func (gmh *GetMetricHandler) HandleGetMetricJSON(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("handle get metric json")

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

	if metrics.MType != models.Gauge && metrics.MType != models.Counter {
		logger.Log.Error("unsupported metric type", zap.String("type", metrics.MType))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	storedMetric, err := gmh.storage.GetMetric(metrics.ID)

	if err != nil {
		logger.Log.Error("metric not found", zap.String("id", metrics.ID))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.Metrics{})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	if err := enc.Encode(storedMetric); err != nil {
		logger.Log.Debug("error encoding response", logger.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
