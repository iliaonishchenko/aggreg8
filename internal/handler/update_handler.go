package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/iliaonishchenko/aggreg8/internal/audit"
	"github.com/iliaonishchenko/aggreg8/internal/logger"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/iliaonishchenko/aggreg8/internal/service"
)

// UpdateHandler обрабатывает HTTP-запросы на создание и обновление метрик.
type UpdateHandler struct {
	storage  service.MetricStorage
	notifier audit.Notifier
}

// NewUpdateHandler создаёт новый UpdateHandler с указанным хранилищем и нотификатором аудита.
func NewUpdateHandler(memStorage service.MetricStorage, notifier audit.Notifier) *UpdateHandler {
	return &UpdateHandler{
		storage:  memStorage,
		notifier: notifier,
	}
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

	auditEvent := audit.AuditEvent{
		TS:        time.Now().Unix(),
		Metrics:   []string{name},
		IPAddress: r.RemoteAddr,
	}
	errs := h.notifier.NotifyAll(auditEvent)
	if len(errs) > 0 {
		for _, err := range errs {
			logger.Log.Error("failed to audit metric", logger.Err(err))
		}
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

	if ok := h.storage.UpdateMetric(&metrics); !ok {
		logger.Log.Error("failed to update metric")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	auditEvent := audit.AuditEvent{
		TS:        time.Now().Unix(),
		Metrics:   []string{metrics.ID},
		IPAddress: r.RemoteAddr,
	}
	errs := h.notifier.NotifyAll(auditEvent)
	if len(errs) > 0 {
		for _, err := range errs {
			logger.Log.Error("failed to audit json metric", logger.Err(err))
		}
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
	err := h.storage.UpdateMetrics(metrics)
	if err != nil {
		logger.Log.Error("failed to update metrics batch", logger.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	metricNames := make([]string, 0, len(metrics))
	for _, metric := range metrics {
		metricNames = append(metricNames, metric.ID)
	}

	auditEvent := audit.AuditEvent{
		TS:        time.Now().Unix(),
		Metrics:   metricNames,
		IPAddress: r.RemoteAddr,
	}
	errs := h.notifier.NotifyAll(auditEvent)
	if len(errs) > 0 {
		for _, err := range errs {
			logger.Log.Error("failed to audit metrics batch", logger.Err(err))
		}
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
