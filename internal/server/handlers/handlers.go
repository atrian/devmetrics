package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/atrian/devmetrics/internal/dto"
	"github.com/atrian/devmetrics/internal/server/storage"
)

type Handler struct {
	*chi.Mux
	storage storage.Repository
}

func NewHandler() *Handler {
	h := &Handler{
		Mux:     chi.NewMux(),
		storage: storage.NewMemoryStorage(),
	}

	// По запросу GET http://<АДРЕС_СЕРВЕРА>/ сервер должен отдавать HTML-страничку со списком имён
	// и значений всех известных ему на текущий момент метрик.
	h.Get("/", h.GetMetrics())

	// Сервер должен возвращать текущее значение запрашиваемой метрики в текстовом виде по запросу
	// GET http://<АДРЕС_СЕРВЕРА>/value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ> (со статусом http.StatusOK).
	// При попытке запроса неизвестной серверу метрики сервер должен возвращать http.StatusNotFound.
	h.Get("/value/{metricType}/{metricTitle}", h.GetMetric())

	// Сохранение произвольных метрик,
	// POST /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	h.Post("/update/{metricType}/{metricTitle}/{metricValue}", h.UpdateMetric())

	h.Post("/value/", h.GetJSONMetric())
	// Пробуем анмаршалинг
	h.Post("/update/", h.UpdateJSONMetric())

	return h
}

// GetMetrics получение всех сохраненных метрик в html формате GET /
func (h *Handler) GetMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		html, err := template.ParseFiles("internal/server/templates/metricTemplate.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		html.Execute(w, h.storage.GetMetrics())
	}
}

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

// GetJSONMetric получение метрик GET /value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>
func (h *Handler) GetJSONMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json") //TODO убрать отсюда работу с заголовками, сделать через конфиг
		metricCandidate := unmarshallMetric(r.Body)

		switch metricCandidate.MType {
		case "gauge":
			if metricValue, exist := h.storage.GetGauge(metricCandidate.ID); exist {
				metricCandidate.Value = &metricValue
				JSONMetric, err := json.Marshal(metricCandidate)

				if err != nil {
					panic(err)
				}

				w.WriteHeader(http.StatusOK)
				w.Write(JSONMetric)
				return
			} else {
				http.Error(w, "gauge not found", http.StatusNotFound)
				return
			}

		case "counter":
			if metricValue, exist := h.storage.GetCounter(metricCandidate.ID); exist {
				metricCandidate.Delta = &metricValue
				JSONMetric, err := json.Marshal(metricCandidate)

				if err != nil {
					panic(err)
				}

				w.WriteHeader(http.StatusOK)
				w.Write(JSONMetric)
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

// UpdateJSONMetric обновление метрик POST /update в JSON
func (h *Handler) UpdateJSONMetric() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric := unmarshallMetric(r.Body)

		switch metric.MType {
		case "gauge":
			h.storage.StoreGauge(metric.ID, *metric.Value)
			currentValue, _ := h.storage.GetGauge(metric.ID)
			metric.Value = &currentValue
		case "counter":
			h.storage.StoreCounter(metric.ID, *metric.Delta)
			currentValue, _ := h.storage.GetCounter(metric.ID)
			metric.Delta = &currentValue
		default:
		}

		w.Header().Set("content-type", "application/json") // TODO убрать?
		// устанавливаем статус-код 200
		w.WriteHeader(http.StatusOK)

		fmt.Println("Request OK", metric)
		json.NewEncoder(w).Encode(metric)
	}
}

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

		w.Header().Set("content-type", "application/json") // TODO убрать?
		// устанавливаем статус-код 200
		w.WriteHeader(http.StatusOK)
		fmt.Println("Request OK")

	}
}

// обновление метрик POST /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
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

func unmarshallMetrics(r *http.Request) []dto.Metrics {

	decoder := json.NewDecoder(r.Body)
	metrics := make([]dto.Metrics, 0, 10)
	err := decoder.Decode(&metrics)
	if err != nil {
		panic(err)
	}

	return metrics // TODO возвращать ошибку если неудалось разобрать JSON
}

func unmarshallMetric(body io.ReadCloser) *dto.Metrics {
	decoder := json.NewDecoder(body)
	var metric dto.Metrics
	err := decoder.Decode(&metric)
	if err != nil {
		panic(err)
	}

	return &metric
}
