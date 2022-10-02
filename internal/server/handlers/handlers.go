package handlers

import (
	"fmt"
	"github.com/atrian/devmetrics/internal/appconfig"
	"github.com/atrian/devmetrics/internal/server/storage"
	"net/http"
	"strconv"
	"strings"
)

type UpdateMetricHandler struct {
	storage storage.Repository
	config  appconfig.Config
}

type metricCandidate struct {
	metricType  string
	metricTitle string
	metricValue string
}

func NewUpdateMetricHandler() *UpdateMetricHandler {
	return &UpdateMetricHandler{storage: storage.NewMemoryStorage()}
}

func (h UpdateMetricHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	// Отлпдочная информация по запросу
	fmt.Println(r.Method)
	fmt.Println(r.URL)

	badRequestFlag := false
	var actualMetricValue string
	metricCandidate, statusCode := validateRequest(r)

	if statusCode != http.StatusOK {
		fmt.Println("Всё пропало")
		http.Error(w, "Can't validate update request", statusCode)
		return
	}

	fmt.Println("------------------")
	fmt.Println(metricCandidate.metricType)
	fmt.Println("------------------")

	if metricCandidate.metricType == "gauge" {
		if res := h.storage.StoreGauge(metricCandidate.metricTitle, metricCandidate.metricValue); res {

			// значение успешно сохранено
			fmt.Printf("Gauge metric %v stored with value %v\n",
				metricCandidate.metricTitle, metricCandidate.metricValue)
			actualMetricValue = metricCandidate.metricValue
		} else {
			badRequestFlag = true
			fmt.Println("Cant store Gauge metric")
		}
	}

	if metricCandidate.metricType == "counter" {
		if res := h.storage.StoreCounter(metricCandidate.metricTitle, metricCandidate.metricValue); res {

			// значение успешно сохранено
			fmt.Printf("Counter metric %v stored. Current value is: %v\n",
				metricCandidate.metricTitle, h.storage.GetCounter(metricCandidate.metricTitle))
			actualMetricValue = strconv.Itoa(int(h.storage.GetCounter(metricCandidate.metricTitle)))
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

// валидируем запрос и в случае если все ок отдаем метрику на сохранение
func validateRequest(r *http.Request) (*metricCandidate, int) {

	endpointParts := endpointParser(r.URL.Path)
	metricCandidate := metricCandidate{}

	if len(endpointParts) < 4 {
		return &metricCandidate, http.StatusNotFound
	}

	// endpointParts[1] ТИП_МЕТРИКИ gauge, counter
	// endpointParts[2] ИМЯ_МЕТРИКИ
	// endpointParts[3] ЗНАЧЕНИЕ_МЕТРИКИ

	if endpointParts[1] != "gauge" && endpointParts[1] != "counter" {
		return &metricCandidate, http.StatusNotImplemented
	}

	metricCandidate.metricType = endpointParts[1]
	metricCandidate.metricTitle = endpointParts[2]
	metricCandidate.metricValue = endpointParts[3]

	return &metricCandidate, http.StatusOK
}

// разбираем URL.Path в слайс по "/"
func endpointParser(endpoint string) []string {
	return strings.FieldsFunc(endpoint, func(r rune) bool {
		return r == '/'
	})
}
