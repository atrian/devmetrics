package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// UpdateMetric обновление метрик POST /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
func (h *Handler) UpdateMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		badRequestFlag := false
		var actualMetricValue string

		metricType := chi.URLParam(r, "metricType")
		metricTitle := chi.URLParam(r, "metricTitle")
		metricValue := chi.URLParam(r, "metricValue")

		if metricType != "gauge" && metricType != "counter" {
			http.Error(w, "Not implemented", http.StatusNotImplemented)
			return
		}

		h.logger.Debug("UpdateMetric chi.URLParam", zap.String("metricType", metricType))

		if metricType == "gauge" {
			floatValue, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				badRequestFlag = true
			}

			err = h.storage.StoreGauge(metricTitle, floatValue)
			if err != nil {
				h.logger.Error("Cant store gauge metric", zap.Error(errors.Unwrap(err)))
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
			actualMetricValue = metricValue

			h.logger.Debug("Gauge metric stored",
				zap.String("metric", metricTitle),
				zap.String("actualMetricValue", actualMetricValue))
		}

		if metricType == "counter" {
			intValue, err := strconv.Atoi(metricValue)
			if err != nil {
				badRequestFlag = true
			}

			err = h.storage.StoreCounter(metricTitle, int64(intValue))
			if err != nil {
				h.logger.Error("Cant store counter metric", zap.Error(errors.Unwrap(err)))
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}

			counterVal, _ := h.storage.GetCounter(metricTitle)
			actualMetricValue = strconv.Itoa(int(counterVal))

			h.logger.Debug("Counter metric stored",
				zap.String("metric", metricTitle),
				zap.String("actualMetricValue", actualMetricValue))
		}

		if badRequestFlag {
			h.logger.Info("Cant store metric, cant convert string value to storage value")
			http.Error(w, "Cant store metric", http.StatusBadRequest)
		} else {
			h.logger.Debug("Request OK")

			w.Header().Set("content-type", "text/plain")
			// устанавливаем статус-код 200
			w.WriteHeader(http.StatusOK)

			_, err := fmt.Fprint(w, actualMetricValue)
			if err != nil {
				h.logger.Error("UpdateMetric handler response writer", zap.Error(err))
			}
		}
	}
}
