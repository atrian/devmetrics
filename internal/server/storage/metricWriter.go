package storage

import (
	"encoding/json"
	"os"

	"github.com/atrian/devmetrics/internal/dto"
)

type metricWriter struct {
	file    *os.File
	encoder *json.Encoder
}

func NewMetricWriter(fileName string) (*metricWriter, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return nil, err
	}
	return &metricWriter{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (mw *metricWriter) WriteMetric(metrics *[]dto.Metrics) error {
	return mw.encoder.Encode(&metrics)
}

func (mw *metricWriter) Close() error {
	return mw.file.Close()
}
