package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/atrian/devmetrics/internal/appconfig/agentconfig"
	"github.com/atrian/devmetrics/internal/crypter"
	"github.com/atrian/devmetrics/internal/dto"
	"github.com/atrian/devmetrics/internal/signature"
	"github.com/atrian/devmetrics/pkg/logger"
	pb "github.com/atrian/devmetrics/proto"
)

// Uploader отправляет данные метрик и счетчиков на удаленный сервер
type Uploader struct {
	HTTPClient     *http.Client        // HTTPClient клиент для HTTP транспорта
	GRPCClient     pb.DevMetricsClient // GRPCClient клиент для GRPC транспорта
	GRPCConnection *grpc.ClientConn    // GRPCConnection GRPC соединение
	config         *agentconfig.Config // config конфигурация приложения
	hasher         signature.Hasher    // hasher подпись метрик
	crypter        crypter.Crypter     // crypter отправка шифрованных данных
	logger         logger.Logger
}

// NewUploader принимает конфигурацию и логгер, подключает зависимости:
// crypto.Sha256Hasher, http.Client
func NewUploader(config *agentconfig.Config, logger logger.Logger) *Uploader {
	keyManager := crypter.New()
	if config.Agent.CryptoKey != "" {
		pubKey, err := keyManager.ReadPublicKey(config.Agent.CryptoKey)
		if err != nil {
			logger.Fatal("Can't load public key", err)
		}
		keyManager.RememberPublicKey(pubKey)
	}

	uploader := Uploader{
		HTTPClient: &http.Client{},
		config:     config,
		hasher:     signature.NewSha256Hasher(),
		crypter:    keyManager,
		logger:     logger,
	}

	// Инициализируем GRPC клиент, если выбран соответствующий протокол
	if config.Transport.Protocol == "grpc" {
		// устанавливаем соединение с сервером
		conn, err := grpc.Dial(config.Transport.AddressGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Fatal("Can't connect GRPC server", err)
		}

		uploader.GRPCConnection = conn
		uploader.GRPCClient = pb.NewDevMetricsClient(conn)
	}

	return &uploader
}

// SendStat отправка одной подписанной метрики на сервер в JSON формате
// Deprecated: метод заменен на массовую отправку через SendAllStats
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
			uploader.logger.Error("SendStat json.Marshal", err)
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
			uploader.logger.Error("SendStat json.Marshal", err)
			continue
		}

		uploader.sendRequest(jsonMetric)
	}
}

// MarshallMetrics маршалит текущие значения метрик данные в JSON
func (uploader *Uploader) marshallMetrics(metrics *MetricsDics) ([]byte, error) {
	exportedMetrics := uploader.signMetrics(metrics)
	jsonMetrics, err := json.Marshal(exportedMetrics)
	return jsonMetrics, err
}

// signMetrics подписывает метрики хешом перед отправкой
func (uploader *Uploader) signMetrics(metrics *MetricsDics) *[]dto.Metrics {
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

	return metrics.exportMetrics(configuredHasher)
}

// encryptData шифрует метрику для передачи по HTTP
func (uploader *Uploader) encryptData(data []byte) ([]byte, error) {
	if uploader.config.Agent.CryptoKey != "" {
		secureData, err := uploader.crypter.Encrypt(data)
		if err != nil {
			uploader.logger.Error("Can't encrypt message", err)
		}
		return secureData, err
	}
	return data, nil
}

// SendAllStats отправка всех метрик на сервер
func (uploader *Uploader) SendAllStats(metrics *MetricsDics) {
	if uploader.config.Transport.Protocol == "grpc" {
		uploader.sendStatsViaGrpc(metrics)
	} else {
		uploader.sendStatsViaHttp(metrics)
	}
}

// sendStatsViaGrpc Отправка статистики по протоколу Grpc.
func (uploader *Uploader) sendStatsViaGrpc(metrics *MetricsDics) {
	var upsertMetricsRequest pb.UpsertMetricsRequest

	// TODO добавить подпись метрик в PROTO?
	exportedMetrics := uploader.signMetrics(metrics)

	for _, metric := range *exportedMetrics {
		switch metric.MType {
		case "gauge":
			upsertMetricsRequest.Metrics = append(upsertMetricsRequest.Metrics, &pb.Metric{
				Type: &pb.Metric_Gauge{
					Gauge: &pb.Gauge{
						ID:    metric.ID,
						Value: *metric.Value,
					},
				},
			})
		case "counter":
			upsertMetricsRequest.Metrics = append(upsertMetricsRequest.Metrics, &pb.Metric{
				Type: &pb.Metric_Counter{
					Counter: &pb.Counter{
						ID:    metric.ID,
						Delta: *metric.Delta,
					},
				},
			})
		default:
		}
	}

	_, err := uploader.GRPCClient.UpdateMetrics(context.Background(), &upsertMetricsRequest)
	if err != nil {
		uploader.logger.Error("GRPCClient.UpdateMetrics failed", err)
	}
}

// sendStatsViaHttp Отправка статистики по протоколу Transport. С шифрованием и сжатием Gzip
func (uploader *Uploader) sendStatsViaHttp(metrics *MetricsDics) {
	// маршалим данные в JSON
	data, err := uploader.marshallMetrics(metrics)
	if err != nil {
		uploader.logger.Error("SendAllStats marshallMetrics error", err)
	}

	// шифруем данные при необходимости
	data, err = uploader.encryptData(data)
	if err != nil {
		uploader.logger.Error("SendAllStats encryptData error", err)
	}

	uploader.sendGzippedRequest(data)
}

// sendRequest отправка запроса, используется для отправки одной метрики методом POST
// без сжатия
func (uploader *Uploader) sendRequest(body []byte) {
	// строим адрес сервера
	endpoint := uploader.buildStatUploadURL()

	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		uploader.logger.Error("sendRequest NewRequest", err)
	}

	// устанавливаем заголовки
	request.Header.Set("Content-Type", uploader.config.Transport.ContentType)

	resp, err := uploader.HTTPClient.Do(request)
	if err != nil {
		uploader.logger.Error("sendRequest HTTPClient.Do", err)
		return
	}

	bcErr := resp.Body.Close()
	if bcErr != nil {
		uploader.logger.Error("sendRequest Body.Close error", bcErr)
	}
}

// sendGzippedRequest отправка запроса, используется для отправки метрик методом POST
// Используется gzip сжатие, передается заголовок Content-Encoding: gzip
func (uploader *Uploader) sendGzippedRequest(body []byte) {
	if len(body) == 0 {
		uploader.logger.Debug("Empty body, return")
		return
	}

	var gzBody bytes.Buffer
	endpoint := uploader.buildStatsUploadURL()

	gzipWriter := gzip.NewWriter(&gzBody)
	if _, err := gzipWriter.Write(body); err != nil {
		uploader.logger.Error("sendGzippedRequest gzipWriter.Write", err)
		return
	}
	err := gzipWriter.Close()
	if err != nil {
		uploader.logger.Error("sendGzippedRequest gzipWriter.Close", err)
		return
	}

	// собираем request
	request, err := http.NewRequest(http.MethodPost, endpoint, &gzBody)
	if err != nil {
		uploader.logger.Error("sendGzippedRequest http.NewRequest", err)
	}

	// устанавливаем заголовки
	request.Header.Set("X-Real-IP", uploader.config.Agent.AgentIP.String())
	request.Header.Set("Content-Type", uploader.config.Transport.ContentType)
	request.Header.Set("Content-Encoding", "gzip")

	resp, err := uploader.HTTPClient.Do(request)
	if err != nil {
		uploader.logger.Error("sendGzippedRequest HTTPClient.Do", err)
		return
	}

	bcErr := resp.Body.Close()
	if bcErr != nil {
		uploader.logger.Error("sendGzippedRequest Body.Close error", err)
	}
}

// buildStatUploadURL построение целевого адреса для отправки одной метрики
// Deprecated: отправка одной метрики больше не используется, применяйте buildStatsUploadURL
func (uploader *Uploader) buildStatUploadURL() string {
	return fmt.Sprintf(uploader.config.Transport.URLTemplate,
		uploader.config.Transport.Protocol,
		uploader.config.Transport.AddressHTTP) + "update/"
}

// buildStatsUploadURL построение целевого адреса для массовой отправки метрик
func (uploader *Uploader) buildStatsUploadURL() string {
	return fmt.Sprintf(uploader.config.Transport.URLTemplate,
		uploader.config.Transport.Protocol,
		uploader.config.Transport.AddressHTTP) + "updates/"
}
