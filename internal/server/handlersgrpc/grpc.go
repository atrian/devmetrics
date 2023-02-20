package handlersgrpc

import (
	"github.com/atrian/devmetrics/internal/server/storage"
	"github.com/atrian/devmetrics/pkg/logger"
	pb "github.com/atrian/devmetrics/proto"
)

type MetricServer struct {
	// pb.UnimplementedDevMetricsServer для совместимости с будущими версиями
	pb.UnimplementedDevMetricsServer
	storage storage.Repository
	logger  logger.Logger
}

func NewMetricServer(storage storage.Repository, logger logger.Logger) *MetricServer {
	ms := MetricServer{
		UnimplementedDevMetricsServer: pb.UnimplementedDevMetricsServer{},
		storage:                       storage,
		logger:                        logger,
	}

	return &ms
}
