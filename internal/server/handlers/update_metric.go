package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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

		fmt.Println("------------------")
		fmt.Println(metricType)
		fmt.Println("------------------")

		if metricType == "gauge" {
			floatValue, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				badRequestFlag = true
			}

			h.storage.StoreGauge(metricTitle, floatValue)
			actualMetricValue = metricValue

			fmt.Printf("Gauge metric %v stored. Current value is: %v\n",
				metricTitle, actualMetricValue)
		}

		if metricType == "counter" {
			intValue, err := strconv.Atoi(metricValue)
			if err != nil {
				badRequestFlag = true
			}

			h.storage.StoreCounter(metricTitle, int64(intValue))

			counterVal, _ := h.storage.GetCounter(metricTitle)
			actualMetricValue = strconv.Itoa(int(counterVal))
			fmt.Printf("Gauge metric %v stored. Current value is: %v\n",
				metricTitle, actualMetricValue)
		}

		if badRequestFlag {
			fmt.Println("Cant store metric")
			http.Error(w, "Cant store metric", http.StatusBadRequest)
		} else {
			fmt.Println("Request OK")

			w.Header().Set("content-type", "text/plain")
			// устанавливаем статус-код 200
			w.WriteHeader(http.StatusOK)

			fmt.Fprint(w, actualMetricValue)
		}
	}
}
