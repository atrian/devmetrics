package agent

import (
	"fmt"
	"net/http"
	"os"
)

type Uploader struct {
	client *http.Client
	config *HTTPConfig
}

func NewUploader(config *HTTPConfig) *Uploader {
	uploader := Uploader{
		client: &http.Client{},
		config: config,
	}
	return &uploader
}

// SendStat отправка метрик на сервер
func (uploader *Uploader) SendStat(metrics *MetricsDics) {
	for key, metric := range metrics.GaugeDict {
		endpoint := uploader.buildStatUploadURL("gauge", key, fmt.Sprintf("%f", metric.value))
		println(endpoint)
		uploader.sendRequest(endpoint)
	}

	for key, metric := range metrics.CounterDict {
		endpoint := uploader.buildStatUploadURL("counter", key, fmt.Sprintf("%d", metric.value))
		println(endpoint)
		uploader.sendRequest(endpoint)
	}
}

// отправка запроса, обработка ответа
func (uploader *Uploader) sendRequest(endpoint string) {
	request, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

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
func (uploader *Uploader) buildStatUploadURL(metricType string, metricTitle string, metricValue string) string {
	return fmt.Sprintf(uploader.config.URLTemplate, uploader.config.Server, uploader.config.Port, metricType, metricTitle, metricValue)
}
