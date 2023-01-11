package storage

import (
	"encoding/json"
	"os"

	"github.com/atrian/devmetrics/internal/dto"
)

// MetricWriter записывает данные из слайса dto.Metrics в файл на диске
type MetricWriter struct {
	file    *os.File
	encoder *json.Encoder
}

// NewMetricWriter принимает на вход имя файла и подготавливает MetricWriter со всеми зависимостями
func NewMetricWriter(fileName string) (*MetricWriter, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return nil, err
	}
	return &MetricWriter{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

// WriteMetric дамп метрик в JSON формате на диск
func (mw *MetricWriter) WriteMetric(metrics *[]dto.Metrics) error {
	return mw.encoder.Encode(&metrics)
}

// Close закрывает файл с дампом, совобождает ресурсы
func (mw *MetricWriter) Close() error {
	return mw.file.Close()
}
