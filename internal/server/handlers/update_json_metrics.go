package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/atrian/devmetrics/internal/dto"
)

// UpdateJSONMetrics обновление метрик POST /updates/ в JSON
//
//	@Tags Metrics
//	@Summary Массовое обновление данных метрик с передачей данных в JSON формате
//	@Accept  json
//	@Produce json
//	@Param metrics body []dto.Metrics true "Принимает JSON массивом метрик, возвращает JSON с обновленными данными"
//	@Success 200 {array} dto.Metrics
//	@Failure 400 {string} string ""
//	@Failure 404 {string} string ""
//	@Failure 500 {string} string ""
//	@Router /updates/ [post]
func (h *Handler) UpdateJSONMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// список уникальных метрик в запросе
		countersRequested := make(map[string]int)
		gaugesRequested := make(map[string]int)

		metrics, err := h.unmarshallMetrics(r)
		if err != nil {
			h.logger.Error("UpdateJSONMetrics cant unmarshallMetric", err)
			http.Error(w, "Bad JSON", http.StatusBadRequest)
		}
		verifiedMetrics := make([]dto.Metrics, 0, len(metrics))

		for _, metric := range metrics {
			switch metric.MType {
			case "gauge":
				// если валидация подписи нужна и она не прошла, пропускаем метрику
				if h.config.Server.HashKey != "" && !h.hasher.Compare(metric.Hash,
					fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value),
					h.config.Server.HashKey) {
					continue
				}
				gaugesRequested[metric.ID] += 1
			case "counter":
				if h.config.Server.HashKey != "" && !h.hasher.Compare(metric.Hash,
					fmt.Sprintf("%s:counter:%d", metric.ID, *metric.Delta),
					h.config.Server.HashKey) {
					continue
				}
				countersRequested[metric.ID] += 1
			default:
				// непонятные метрики просто пропускаем
				continue
			}
			verifiedMetrics = append(verifiedMetrics, metric)
		}

		// сохраняем метрики с правильными подписями в БД
		h.storage.SetMetrics(verifiedMetrics)

		// слайс уникальных метрик для ответа с актуальными значениями
		responseMetrics := make([]dto.Metrics, 0, len(countersRequested)+len(gaugesRequested))

		// собираем актуальные значения counters
		for key := range countersRequested {
			actualCounterValue, _ := h.storage.GetCounter(key)

			metric := dto.Metrics{
				ID:    key,
				MType: "counter",
				Delta: &actualCounterValue,
			}

			// подписываем метрику если установлен ключ шифрования
			if h.config.Server.HashKey != "" {
				metric.Hash = h.hasher.Hash(fmt.Sprintf("%s:counter:%d", metric.ID, actualCounterValue),
					h.config.Server.HashKey)
			}

			responseMetrics = append(responseMetrics, metric)
		}

		// собираем актуальные значения gauges
		for key := range gaugesRequested {
			actualGaugeValue, _ := h.storage.GetGauge(key)

			metric := dto.Metrics{
				ID:    key,
				MType: "gauge",
				Value: &actualGaugeValue,
			}

			// подписываем метрику если установлен ключ шифрования
			if h.config.Server.HashKey != "" {
				metric.Hash = h.hasher.Hash(fmt.Sprintf("%s:gauge:%f", metric.ID, actualGaugeValue),
					h.config.Server.HashKey)
			}

			responseMetrics = append(responseMetrics, metric)
		}

		w.Header().Set("content-type", h.config.Transport.ContentType)
		// устанавливаем статус-код 200
		w.WriteHeader(http.StatusOK)
		h.logger.Debug("Request OK")

		// формируем структуру JSON ответа
		response := struct {
			Status  string
			Updated []dto.Metrics
		}{
			Status:  "OK",
			Updated: responseMetrics,
		}

		jeErr := json.NewEncoder(w).Encode(response)
		if jeErr != nil {
			h.logger.Error("json.NewEncoder err", jeErr)
		}
	}
}

// unmarshallMetrics анмаршаллинг метрик в слайс dto.Metrics
func (h *Handler) unmarshallMetrics(r *http.Request) ([]dto.Metrics, error) {
	var body io.Reader

	// если в заголовках установлен Content-Encoding gzip, распаковываем тело
	if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
		body = h.decodeGzipBody(r.Body)
	} else {
		body = r.Body
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			h.logger.Error("Body io.ReadCloser error", err)
		}
	}(r.Body)

	// если в настройках установлен ключ шифрования, расшифровываем метрики
	if h.crypter.ReadyForDecrypt() {
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(body)
		if err != nil {
			h.logger.Error("Read from r.body failed", err)
		}

		message, err := h.decryptMessage(buf.Bytes())
		if err != nil {
			h.logger.Error("decryptMessage failed", err)
		}

		body = bytes.NewReader(message)
	}

	var metrics []dto.Metrics
	decoder := json.NewDecoder(body)
	err := decoder.Decode(&metrics)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

// decodeGzipBody распаковка GZIP тела запроса
func (h *Handler) decodeGzipBody(gzipR io.Reader) io.Reader {
	gz, err := gzip.NewReader(gzipR)
	if err != nil {
		h.logger.Error("decodeGzipBody cant set up gzip decoder", err)
	}
	return gz
}

func (h *Handler) decryptMessage(encryptedMessage []byte) ([]byte, error) {
	message, err := h.crypter.Decrypt(encryptedMessage)
	if err != nil {
		return nil, err
	}
	return message, nil
}
