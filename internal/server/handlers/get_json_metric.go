package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GetJSONMetric получение метрик POST /value/
//
//	@Tags Metrics
//	@Summary Запрос одной метрики с указанием её типа и имени
//	@Accept  json
//	@Produce json
//	@Param metric body dto.EmptyMetric true "Сервис принимает пустую метрику с указанием типа и имени метрики, отдает JSON наполненный данными"
//	@Success 200 {object} dto.Metrics
//	@Failure 400 {string} string ""
//	@Failure 404 {string} string ""
//	@Failure 500 {string} string ""
//	@Router /value/ [post]
func (h *Handler) GetJSONMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", h.config.HTTP.ContentType)
		metricCandidate, err := h.unmarshallMetric(r.Body)

		if err != nil {
			h.logger.Error("GetJSONMetric cant unmarshallMetric", err)
			http.Error(w, "Bad JSON", http.StatusBadRequest)
		}

		switch metricCandidate.MType {
		case "gauge":
			if metricValue, exist := h.storage.GetGauge(metricCandidate.ID); exist {

				// подписываем метрику если установлен ключ шифрования
				if h.config.Server.HashKey != "" {
					metricCandidate.Hash = h.hasher.Hash(fmt.Sprintf("%s:gauge:%f", metricCandidate.ID, metricValue),
						h.config.Server.HashKey)
				}

				metricCandidate.Value = &metricValue
				JSONMetric, err := json.Marshal(metricCandidate)

				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}

				w.WriteHeader(http.StatusOK)
				_, wErr := w.Write(JSONMetric)
				if wErr != nil {
					http.Error(w, wErr.Error(), http.StatusInternalServerError)
				}
				return
			} else {
				http.Error(w, "gauge not found", http.StatusNotFound)
				return
			}

		case "counter":
			if metricValue, exist := h.storage.GetCounter(metricCandidate.ID); exist {

				// подписываем метрику если установлен ключ шифрования
				if h.config.Server.HashKey != "" {
					metricCandidate.Hash = h.hasher.Hash(fmt.Sprintf("%s:counter:%d", metricCandidate.ID, metricValue),
						h.config.Server.HashKey)
				}

				metricCandidate.Delta = &metricValue
				JSONMetric, err := json.Marshal(metricCandidate)

				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}

				w.WriteHeader(http.StatusOK)
				_, wErr := w.Write(JSONMetric)
				if wErr != nil {
					http.Error(w, wErr.Error(), http.StatusInternalServerError)
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
