package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// GetMetric получение метрик GET /value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>
func (h *Handler) GetMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricTitle := chi.URLParam(r, "metricTitle")

		switch metricType {
		case "gauge":
			if metricValue, exist := h.storage.GetGauge(metricTitle); exist {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf("%v", metricValue)))
				return
			} else {
				http.Error(w, "gauge not found", http.StatusNotFound)
				return
			}

		case "counter":
			if metricValue, exist := h.storage.GetCounter(metricTitle); exist {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf("%v", metricValue)))
				return
			} else {
				http.Error(w, "counter not found", http.StatusNotFound)
				return
			}

		default:
			http.Error(w, "Not implemented", http.StatusNotImplemented)
			return
		}
	}
}
