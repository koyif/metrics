package server

import (
	"context"
	"time"

	"github.com/koyif/metrics/internal/audit"
	"github.com/koyif/metrics/internal/config"
	"github.com/koyif/metrics/internal/grpc/converter"
	"github.com/koyif/metrics/internal/grpc/interceptor"
	"github.com/koyif/metrics/internal/models"
	"github.com/koyif/metrics/internal/proto/api/proto"
	"github.com/koyif/metrics/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	metricIDEmptyErrorMessage          = "metric ID cannot be empty"
	emptyMetricsErrorMessage           = "metrics array cannot be empty"
	failedToPersistMetricsErrorMessage = "failed to persist metrics"
)

type metricsStorer interface {
	StoreAll(metrics []models.Metrics) error
	Persist() error
}

// MetricsServer implements the gRPC Metrics service.
type MetricsServer struct {
	proto.UnimplementedMetricsServer
	service      metricsStorer
	cfg          *config.Config
	auditManager *audit.Manager
}

// NewMetricsServer creates a new gRPC metrics server.
// The auditManager can be nil if auditing is not enabled.
func NewMetricsServer(service metricsStorer, cfg *config.Config, auditManager *audit.Manager) *MetricsServer {
	return &MetricsServer{
		service:      service,
		cfg:          cfg,
		auditManager: auditManager,
	}
}

// UpdateMetrics implements the gRPC UpdateMetrics RPC method.
// It receives a batch of metrics from the agent, validates them, and stores them.
func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	if len(req.Metrics) == 0 {
		logger.Log.Warn(emptyMetricsErrorMessage)
		return nil, status.Error(codes.InvalidArgument, emptyMetricsErrorMessage)
	}

	for _, metric := range req.Metrics {
		if metric.Id == "" {
			logger.Log.Warn(metricIDEmptyErrorMessage)
			return nil, status.Error(codes.InvalidArgument, metricIDEmptyErrorMessage)
		}
	}

	metrics := converter.ProtoToModels(req.Metrics)

	if err := s.service.StoreAll(metrics); err != nil {
		logger.Log.Warn(failedToPersistMetricsErrorMessage, logger.Error(err))
		return nil, status.Error(codes.Internal, failedToPersistMetricsErrorMessage)
	}

	if s.cfg.StoreInterval.Value() == 0 {
		if err := s.service.Persist(); err != nil {
			logger.Log.Warn(failedToPersistMetricsErrorMessage, logger.Error(err))
			return nil, status.Error(codes.Internal, failedToPersistMetricsErrorMessage)
		}
	}

	if s.auditManager != nil {
		metricNames := make([]string, 0, len(metrics))
		for _, metric := range metrics {
			metricNames = append(metricNames, metric.ID)
		}

		clientIP, err := interceptor.ExtractIPFromMetadata(ctx)
		if err != nil {
			clientIP = "unknown"
		}

		s.sendAuditEvent(metricNames, clientIP)
	}

	return &proto.UpdateMetricsResponse{}, nil
}

// sendAuditEvent sends an audit event for the stored metrics.
func (s *MetricsServer) sendAuditEvent(metricNames []string, clientIP string) {
	if s.auditManager == nil || !s.auditManager.IsEnabled() {
		return
	}

	event := models.AuditEvent{
		Timestamp: time.Now().Unix(),
		Metrics:   metricNames,
		IPAddress: clientIP,
	}

	s.auditManager.NotifyAll(event)
}
