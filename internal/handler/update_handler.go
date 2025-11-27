package handler

import (
	"fmt"
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

func (uh UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricType := r.PathValue("type")
	name := r.PathValue("name")
	value := r.PathValue("value")
	
	fmt.Printf("received metric: type: %s, name: %s, value: %v", metricType, name, value)

	switch metricType {
	case models.Gauge:
		parsedValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = uh.memStorage.UpdateGauge(name, parsedValue)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case models.Counter:
		parsedValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = uh.memStorage.UpdateCounter(name, parsedValue)
		if err != nil {
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
