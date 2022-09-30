package handlers

import (
	"errors"
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
	metricCandidate, err := validateRequest(r)

	if err != nil {
		fmt.Println("Всё пропало")
		http.Error(w, "Can't validate update request", http.StatusBadRequest)
		return
	}

	fmt.Println("------------------")
	fmt.Println(metricCandidate.metricType)
	fmt.Println("------------------")

	if metricCandidate.metricType == "gauge" {
		if res := h.storage.StoreGauge(metricCandidate.metricTitle, metricCandidate.metricValue); res == true {

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
		if res := h.storage.StoreCounter(metricCandidate.metricTitle, metricCandidate.metricValue); res == true {

			// значение успешно сохранено
			fmt.Printf("Counter metric %v stored. Current value is: %v\n",
				metricCandidate.metricTitle, h.storage.GetCounter(metricCandidate.metricTitle))
			actualMetricValue = strconv.Itoa(int(h.storage.GetCounter(metricCandidate.metricTitle)))
		} else {
			badRequestFlag = true
			fmt.Println("Cant store Counter metric")
		}
	}

	if badRequestFlag == true {
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
func validateRequest(r *http.Request) (*metricCandidate, error) {

	endpointParts := endpointParser(r.URL.Path)
	metricCandidate := metricCandidate{}

	if len(endpointParts) < 4 {
		return &metricCandidate, errors.New("Bad request")
	}

	// endpointParts[1] ТИП_МЕТРИКИ gauge, counter
	// endpointParts[2] ИМЯ_МЕТРИКИ
	// endpointParts[3] ЗНАЧЕНИЕ_МЕТРИКИ

	if endpointParts[1] != "gauge" && endpointParts[1] != "counter" {
		return &metricCandidate, errors.New("Bad request")
	}

	metricCandidate.metricType = endpointParts[1]
	metricCandidate.metricTitle = endpointParts[2]
	metricCandidate.metricValue = endpointParts[3]

	return &metricCandidate, nil
}

// разбираем URL.Path в слайс по "/"
func endpointParser(endpoint string) []string {
	return strings.FieldsFunc(endpoint, func(r rune) bool {
		if r == '/' {
			return true
		}
		return false
	})
}
