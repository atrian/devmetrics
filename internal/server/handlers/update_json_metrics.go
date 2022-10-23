package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/atrian/devmetrics/internal/dto"
	"log"
	"net/http"
)

// UpdateJSONMetrics обновление метрик POST /update в JSON
func (h *Handler) UpdateJSONMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := unmarshallMetrics(r)

		for key, metricCandidate := range metrics {
			fmt.Println("------------------")
			log.Println(key, metricCandidate)

			switch metricCandidate.MType {
			case "gauge":
				h.storage.StoreGauge(metricCandidate.ID, *metricCandidate.Value)
			case "counter":
				h.storage.StoreCounter(metricCandidate.ID, *metricCandidate.Delta)
			default:
			}
		}

		w.Header().Set("content-type", h.config.HTTP.ContentType)
		// устанавливаем статус-код 200
		w.WriteHeader(http.StatusOK)
		fmt.Println("Request OK")
	}
}

func unmarshallMetrics(r *http.Request) []dto.Metrics {
	var metrics []dto.Metrics
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&metrics)
	if err != nil {
		panic(err)
	}

	return metrics // TODO возвращать ошибку если неудалось разобрать JSON
}
