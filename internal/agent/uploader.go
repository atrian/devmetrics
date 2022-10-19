package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/atrian/devmetrics/internal/appconfig"
	"github.com/atrian/devmetrics/internal/dto"
)

type Uploader struct {
	client *http.Client
	config *appconfig.HTTPConfig
}

func NewUploader(config *appconfig.HTTPConfig) *Uploader {
	uploader := Uploader{
		client: &http.Client{},
		config: config,
	}
	return &uploader
}

// SendStat отправка метрик на сервер
func (uploader *Uploader) SendStat(metrics *MetricsDics) {
	for key, metric := range metrics.GaugeDict {
		jsonMetric, err := json.Marshal(&dto.Metrics{
			ID:    key,
			MType: "gauge",
			Delta: nil,
			Value: metric.getGaugeValue(),
		})

		if err != nil {
			log.Fatal("Can't get marshal metric")
		}

		uploader.sendRequest(jsonMetric)
	}

	for key, metric := range metrics.CounterDict {
		jsonMetric, err := json.Marshal(&dto.Metrics{
			ID:    key,
			MType: "counter",
			Delta: metric.getCounterValue(),
			Value: nil,
		})

		if err != nil {
			log.Fatal("Can't get marshal metric")
		}

		uploader.sendRequest(jsonMetric)
	}
}

// отправка запроса, обработка ответа
func (uploader *Uploader) sendRequest(body []byte) {

	// строим адрес сервера
	endpoint := uploader.buildStatUploadURL()
	fmt.Println(endpoint)

	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// устанавливаем заголовки
	request.Header.Set("Content-Type", uploader.config.ContentType)

	response, err := uploader.client.Do(request)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// go vet - response body must be closed
	defer response.Body.Close()
}

// построение целевого адреса для отправки метрики
func (uploader *Uploader) buildStatUploadURL() string {
	return fmt.Sprintf(uploader.config.URLTemplate,
		uploader.config.Protocol,
		uploader.config.Server,
		uploader.config.Port)
}
