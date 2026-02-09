package handler

import (
	"fmt"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/iliaonishchenko/aggreg8/internal/service"
	"net/http"
	"strings"
)

type AllMetricsHandler struct {
	storage service.MetricStorage
}

func NewAllMetricsHandler(s service.MetricStorage) *AllMetricsHandler {
	return &AllMetricsHandler{
		storage: s,
	}
}

func (amh *AllMetricsHandler) HandleAll(w http.ResponseWriter, r *http.Request) {
	metrics := amh.storage.GetAllMetrics()

	var html strings.Builder
	html.WriteString("<html><head><title>metrics</title></head><body>")
	html.WriteString("<h1>All metrics</h1>")
	html.WriteString("<table border='1' cellpadding='10'>")
	html.WriteString("<tr><th>Name</th><th>Type</th><th>Value</th></tr>")

	for _, metric := range metrics {
		html.WriteString("<tr>")
		html.WriteString(fmt.Sprintf("<td>%s</td>", metric.ID))
		html.WriteString(fmt.Sprintf("<td>%s</td>", metric.MType))

		if metric.MType == models.Gauge && metric.Value != nil {
			html.WriteString(fmt.Sprintf("<td>%g</td>", *metric.Value))
		} else if metric.MType == models.Counter && metric.Delta != nil {
			html.WriteString(fmt.Sprintf("<td>%d</td>", *metric.Delta))
		} else {
			html.WriteString("<td>-</td>")
		}

		html.WriteString("</tr>")
	}

	html.WriteString("</table></body></html>")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html.String()))
}
