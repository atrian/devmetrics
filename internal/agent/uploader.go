package agent

import (
	"fmt"
	"net/http"
)

type Uploader struct {
	client *http.Client
	config *HttpConfig
}

func NewUploader(config *HttpConfig) *Uploader {
	uploader := Uploader{
		client: &http.Client{},
		config: config,
	}
	return &uploader
}

// SendStat отправка метрик на сервер
func (uploader *Uploader) SendStat(metrics *MetricsDics) {
	for key, metric := range metrics.GaugeDict {
		endpoint := uploader.buildStatUploadUrl("gauge", key, fmt.Sprintf("%f", metric.value))
		println(endpoint)
		uploader.sendRequest(endpoint)
	}

	for key, metric := range metrics.CounterDict {
		endpoint := uploader.buildStatUploadUrl("counter", key, fmt.Sprintf("%d", metric.value))
		println(endpoint)
		uploader.sendRequest(endpoint)
	}
}

// отправка запроса, обработка ответа
func (uploader *Uploader) sendRequest(endpoint string) {
	request, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		fmt.Println(err)
		//os.Exit(1)
	}

	response, err := uploader.client.Do(request)
	if err != nil {
		fmt.Println(err)
		//os.Exit(1)
	}
	_ = response
}

// построение целевого адреса для отправки метрики
func (uploader *Uploader) buildStatUploadUrl(metricType string, metricTitle string, metricValue string) string {
	return fmt.Sprintf(uploader.config.UrlTemplate, uploader.config.Server, uploader.config.Port, metricType, metricTitle, metricValue)
}
