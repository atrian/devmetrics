package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/atrian/devmetrics/internal/appconfig/agentconfig"
	"github.com/atrian/devmetrics/internal/crypto"
	"github.com/atrian/devmetrics/internal/dto"
)

type Uploader struct {
	client *http.Client
	config *agentconfig.Config
	hasher crypto.Hasher
}

func NewUploader(config *agentconfig.Config) *Uploader {
	uploader := Uploader{
		client: &http.Client{},
		config: config,
		hasher: crypto.NewSha256Hasher(),
	}
	return &uploader
}

// SendStat отправка метрик на сервер
func (uploader *Uploader) SendStat(metrics *MetricsDics) {
	for key, metric := range metrics.GaugeDict {
		gaugeValue := metric.getGaugeValue()
		jsonMetric, err := json.Marshal(&dto.Metrics{
			ID:    key,
			MType: "gauge",
			Delta: nil,
			Value: &gaugeValue,
			Hash:  uploader.hasher.Hash(fmt.Sprintf("%s:gauge:%f", key, gaugeValue), uploader.config.Agent.HashKey),
		})

		if err != nil {
			fmt.Println(err)
			continue
		}

		uploader.sendRequest(jsonMetric)
	}

	for key, metric := range metrics.CounterDict {
		counterValue := metric.getCounterValue()
		jsonMetric, err := json.Marshal(&dto.Metrics{
			ID:    key,
			MType: "counter",
			Delta: &counterValue,
			Value: nil,
			Hash:  uploader.hasher.Hash(fmt.Sprintf("%s:counter:%d", key, counterValue), uploader.config.Agent.HashKey),
		})

		if err != nil {
			fmt.Println(err)
			continue
		}

		uploader.sendRequest(jsonMetric)
	}
}

func (uploader *Uploader) SendAllStats(metrics *MetricsDics) {
	exportedMetrics := metrics.exportMetrics()
	jsonMetrics, err := json.Marshal(exportedMetrics)

	if err != nil {
		fmt.Println("can't marshal metrics to JSON")
		return
	}
	uploader.sendRequest(jsonMetrics)
}

// отправка запроса, обработка ответа
func (uploader *Uploader) sendRequest(body []byte) {

	fmt.Println("sendRequest:", string(body))
	// строим адрес сервера
	endpoint := uploader.buildStatUploadURL()

	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println(err)
		//os.Exit(1)
	}

	// устанавливаем заголовки
	request.Header.Set("Content-Type", uploader.config.HTTP.ContentType)

	resp, err := uploader.client.Do(request)
	if err != nil {
		fmt.Println(err)
		//os.Exit(1)
	}

	if resp != nil {
		defer resp.Body.Close()
	}
}

// построение целевого адреса для отправки метрики
func (uploader *Uploader) buildStatUploadURL() string {
	return fmt.Sprintf(uploader.config.HTTP.URLTemplate,
		uploader.config.HTTP.Protocol,
		uploader.config.HTTP.Address) + "update/"
}
