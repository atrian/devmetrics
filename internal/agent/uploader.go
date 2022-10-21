package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/atrian/devmetrics/internal/appconfig"
	"github.com/atrian/devmetrics/internal/dto"
	"log"
	"net/http"
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
		fmt.Println("Gauge:", key, metric)
		gaugeValue := metric.getGaugeValue()
		jsonMetric, err := json.Marshal(&dto.Metrics{
			ID:    key,
			MType: "gauge",
			Delta: nil,
			Value: &gaugeValue,
		})

		if err != nil {
			panic(err)
		}

		uploader.sendRequest(jsonMetric)
	}

	for key, metric := range metrics.CounterDict {
		fmt.Println("Counter:", key, metric)
		counterValue := metric.getCounterValue()
		jsonMetric, err := json.Marshal(&dto.Metrics{
			ID:    key,
			MType: "counter",
			Delta: &counterValue,
			Value: nil,
		})

		if err != nil {
			panic(err)
		}

		uploader.sendRequest(jsonMetric)
	}
}

func (uploader *Uploader) SendAllStats(metrics *MetricsDics) {
	exportedMetrics := metrics.exportMetrics()
	fmt.Println(exportedMetrics)

	jsonMetrics, err := json.Marshal(exportedMetrics)

	if err != nil {
		log.Fatal("can't marshal metrics to JSON")
	}

	fmt.Println(string(jsonMetrics))
	uploader.sendRequest(jsonMetrics)
}

// отправка запроса, обработка ответа
func (uploader *Uploader) sendRequest(body []byte) {

	fmt.Println("sendRequest:", string(body))
	// строим адрес сервера
	endpoint := uploader.buildStatUploadURL()
	fmt.Println(endpoint)

	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println(err)
		//os.Exit(1)
	}

	// устанавливаем заголовки
	request.Header.Set("Content-Type", uploader.config.ContentType)

	resp, err := uploader.client.Do(request)
	if err != nil {
		fmt.Println(err)
		//os.Exit(1)
	}

	fmt.Println(resp)

	if resp != nil {
		defer resp.Body.Close()
	}
}

// построение целевого адреса для отправки метрики
func (uploader *Uploader) buildStatUploadURL() string {
	return fmt.Sprintf(uploader.config.URLTemplate,
		uploader.config.Protocol,
		uploader.config.Address) + "update/"
}
