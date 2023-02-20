package grpc_handlers

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/atrian/devmetrics/internal/dto"
	pb "github.com/atrian/devmetrics/proto"
)

func (ms *MetricServer) UpdateMetrics(ctx context.Context, in *pb.UpsertMetricsRequest) (*pb.UpsertMetricsResponse, error) {
	var response pb.UpsertMetricsResponse

	metricsSize := len(in.Metrics)
	if metricsSize == 0 {
		return nil, status.Errorf(codes.DataLoss, "Empty request")
	}

	metrics := make([]dto.Metrics, 0, metricsSize)
	ms.logger.Debug(fmt.Sprintf("GRPC request with %v metrics", metricsSize))

	for _, metricCandidate := range in.Metrics {
		switch metricCandidate.Type.(type) {
		case *pb.Metric_Gauge:
			value := metricCandidate.GetGauge().Value
			metrics = append(metrics, dto.Metrics{
				ID:    metricCandidate.GetGauge().ID,
				MType: "gauge",
				Value: &value,
			})
		case *pb.Metric_Counter:
			value := metricCandidate.GetCounter().Delta
			metrics = append(metrics, dto.Metrics{
				ID:    metricCandidate.GetCounter().ID,
				MType: "counter",
				Delta: &value,
			})
		default:
			return nil, status.Errorf(codes.InvalidArgument, "Unsupported metric type")
		}
	}

	ms.storage.SetMetrics(metrics)
	response.Status = pb.UpsertMetricsResponse_OK
	return &response, nil
}
