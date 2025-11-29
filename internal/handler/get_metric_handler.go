package handler

import (
	"github.com/go-chi/chi/v5"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/iliaonishchenko/aggreg8/internal/service"
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

	if mType != models.Gauge && mType != models.Counter {
		http.Error(w, "unknown metric type: "+mType, http.StatusNotFound)
		return
	}

	metric, err := gmh.storage.GetMetric(name)
	if err != nil {
		http.Error(w, "unknown metric: "+name, http.StatusNotFound)
		return
	}

	if metric.Value == nil {
		http.Error(w, "invalid metric data", http.StatusInternalServerError)
		return
	}

	var resultValue string
	if metric.MType == models.Gauge {
		resultValue = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	} else {
		resultValue = strconv.FormatInt(int64(*metric.Value), 10)
	}

	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, resultValue)
}
