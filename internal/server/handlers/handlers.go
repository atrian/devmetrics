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

	h.Post("/value", h.GetJSONMetric())
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

		r.Header.Set("Content-Type", "application/json") //TODO убрать отсюда работу с заголовками, сделать через конфиг

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

		badRequestFlag := false
		actualMetricValue := ""

		metricCandidate := unmarshallMetric(r.Body)

		fmt.Println("------------------")
		log.Println(metricCandidate)
		fmt.Println("------------------")

		if metricCandidate.MType == "gauge" {
			if res := h.storage.StoreGauge(metricCandidate.ID, fmt.Sprintf("%f", *metricCandidate.Value)); res {
				actualMetricValue = fmt.Sprintf("%f", *metricCandidate.Value)
				// значение успешно сохранено
				fmt.Printf("Gauge metric %v stored with value %v\n",
					metricCandidate.ID, *metricCandidate.Value)
			} else {
				badRequestFlag = true
				fmt.Println("Cant store Gauge metric")
			}
		}

		if metricCandidate.MType == "counter" {
			if res := h.storage.StoreCounter(metricCandidate.ID, fmt.Sprintf("%v", *metricCandidate.Delta)); res {

				// значение успешно сохранено
				counterVal, _ := h.storage.GetCounter(metricCandidate.ID)
				fmt.Printf("Counter metric %v stored. Current value is: %v\n",
					metricCandidate.ID, counterVal)
				actualMetricValue = strconv.Itoa(int(counterVal))
			} else {
				badRequestFlag = true
				fmt.Println("Cant store Counter metric")
			}
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
			if res := h.storage.StoreGauge(metricTitle, metricValue); res {

				// значение успешно сохранено
				fmt.Printf("Gauge metric %v stored with value %v\n",
					metricTitle, metricValue)
				actualMetricValue = metricValue
			} else {
				badRequestFlag = true
				fmt.Println("Cant store Gauge metric")
			}
		}

		if metricType == "counter" {
			if res := h.storage.StoreCounter(metricTitle, metricValue); res {

				// значение успешно сохранено
				counterVal, _ := h.storage.GetCounter(metricTitle)
				fmt.Printf("Counter metric %v stored. Current value is: %v\n",
					metricTitle, counterVal)
				actualMetricValue = strconv.Itoa(int(counterVal))
			} else {
				badRequestFlag = true
				fmt.Println("Cant store Counter metric")
			}
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

func unmarshallMetric(body io.ReadCloser) *dto.Metrics {
	decoder := json.NewDecoder(body)
	var metricCandidate dto.Metrics
	err := decoder.Decode(&metricCandidate)
	if err != nil {
		panic(err)
	}

	return &metricCandidate
}
