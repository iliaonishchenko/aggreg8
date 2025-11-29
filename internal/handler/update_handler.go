package handler

import (
	"fmt"
	"github.com/go-chi/chi/v5"
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

	fmt.Printf("received metric: type: %s, name: %s, value: %v", metricType, name, value)

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
		floatValue := float64(parsedValue)
		delta := int64(1)

		metric := models.Metrics{ID: name, MType: models.Counter, Delta: &delta, Value: &floatValue}
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
