package agent

import (
	"bytes"
	"compress/gzip"
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
	// создаем функцию-декоратор для того чтобы не тащить хешер и конфиг в другой слой приложения напрямую
	configuredHasher := func(metricType, id string, delta *int64, value *float64) string {
		switch metricType {
		case "counter":
			return uploader.hasher.Hash(fmt.Sprintf("%s:counter:%d", id, *delta),
				uploader.config.Agent.HashKey)
		case "gauge":
			return uploader.hasher.Hash(fmt.Sprintf("%s:gauge:%f", id, *value),
				uploader.config.Agent.HashKey)
		default:
			return ""
		}
	}

	exportedMetrics := metrics.exportMetrics(configuredHasher)

	jsonMetrics, err := json.Marshal(exportedMetrics)

	if err != nil {
		fmt.Println("can't marshal metrics to JSON")
		return
	}

	uploader.sendGzippedRequest(jsonMetrics)
}

// отправка запроса, обработка ответа
func (uploader *Uploader) sendRequest(body []byte) {
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

func (uploader *Uploader) sendGzippedRequest(body []byte) {
	if len(body) == 0 {
		fmt.Println("Empty body, return")
		return
	}

	var gzBody bytes.Buffer
	endpoint := uploader.buildStatsUploadURL()

	gzipWriter := gzip.NewWriter(&gzBody)
	if _, err := gzipWriter.Write(body); err != nil {
		fmt.Println("gzipWriter.Write error: ", err.Error())
		return
	}
	err := gzipWriter.Close()
	if err != nil {
		fmt.Println("gzipWriter.Close error: ", err.Error())
		return
	}

	// собираем request
	request, err := http.NewRequest(http.MethodPost, endpoint, &gzBody)
	if err != nil {
		fmt.Println(err)
		//os.Exit(1)
	}

	// устанавливаем заголовки
	request.Header.Set("Content-Type", uploader.config.HTTP.ContentType)
	request.Header.Set("Content-Encoding", "gzip")

	resp, err := uploader.client.Do(request)
	if err != nil {
		fmt.Println(err)
	}

	if resp != nil {
		defer resp.Body.Close()
	}
}

// построение целевого адреса для отправки одной метрики
func (uploader *Uploader) buildStatUploadURL() string {
	return fmt.Sprintf(uploader.config.HTTP.URLTemplate,
		uploader.config.HTTP.Protocol,
		uploader.config.HTTP.Address) + "update/"
}

// построение целевого адреса для отправки метрик
func (uploader *Uploader) buildStatsUploadURL() string {
	return fmt.Sprintf(uploader.config.HTTP.URLTemplate,
		uploader.config.HTTP.Protocol,
		uploader.config.HTTP.Address) + "updates/"
}
