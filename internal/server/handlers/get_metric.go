package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// GetMetric получение метрик GET /value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>
// @Tags Metrics
// @Summary Запрос одной метрики с указанием её типа и имени
// @Produce json
// @Param metric_type path string true "Тип метрики: counter, gauge"
// @Param metric_name path string true "Имя метрики"
// @Success 200 {object} dto.Metrics
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /value/{metric_type}/{metric_name} [get]
func (h *Handler) GetMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricTitle := chi.URLParam(r, "metricTitle")

		switch metricType {
		case "gauge":
			if metricValue, exist := h.storage.GetGauge(metricTitle); exist {
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprintf(w, "%v", metricValue)
				if err != nil {
					http.Error(w, "Can't write metric value", http.StatusInternalServerError)
				}
				return
			} else {
				http.Error(w, "gauge not found", http.StatusNotFound)
				return
			}

		case "counter":
			if metricValue, exist := h.storage.GetCounter(metricTitle); exist {
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprintf(w, "%v", metricValue)
				if err != nil {
					http.Error(w, "Can't write metric value", http.StatusInternalServerError)
				}
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
