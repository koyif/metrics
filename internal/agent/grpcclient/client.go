package grpcclient

import (
	"context"
	"fmt"

	"github.com/koyif/metrics/internal/agent/config"
	"github.com/koyif/metrics/internal/grpc/converter"
	"github.com/koyif/metrics/internal/models"
	"github.com/koyif/metrics/internal/proto/api/proto"
	"github.com/koyif/metrics/pkg/logger"
	"github.com/koyif/metrics/pkg/netutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// GRPCMetricsClient implements the metrics client interface for gRPC transport.
type GRPCMetricsClient struct {
	conn    *grpc.ClientConn
	client  proto.MetricsClient
	localIP string
	cfg     *config.Config
}

// New creates a new gRPC metrics client.
// It detects the local IP address and establishes a connection to the gRPC server.
func New(cfg *config.Config) (*GRPCMetricsClient, error) {
	localIP, err := netutil.GetOutboundIP()
	if err != nil {
		return nil, fmt.Errorf("failed to detect local IP address: %w", err)
	}
	logger.Log.Info("detected local IP address", logger.String("ip", localIP))

	conn, err := grpc.NewClient(
		cfg.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	client := proto.NewMetricsClient(conn)

	logger.Log.Info("gRPC client initialized", logger.String("server", cfg.Addr))

	return &GRPCMetricsClient{
		conn:    conn,
		client:  client,
		localIP: localIP,
		cfg:     cfg,
	}, nil
}

// SendMetrics sends a batch of metrics to the gRPC server.
// It adds the local IP address to the request metadata and converts the metrics to proto format.
func (c *GRPCMetricsClient) SendMetrics(metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	protoMetrics := converter.ModelsToProto(metrics)

	ctx := metadata.AppendToOutgoingContext(
		context.Background(),
		"x-real-ip", c.localIP,
	)

	req := &proto.UpdateMetricsRequest{
		Metrics: protoMetrics,
	}

	_, err := c.client.UpdateMetrics(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send metrics via gRPC: %w", err)
	}

	logger.Log.Debug("successfully sent metrics via gRPC", logger.Int("count", len(metrics)))
	return nil
}

// SendMetric sends a single metric to the gRPC server.
// It wraps the metric in a slice and calls SendMetrics.
func (c *GRPCMetricsClient) SendMetric(metric models.Metrics) error {
	return c.SendMetrics([]models.Metrics{metric})
}

// Close closes the gRPC connection.
func (c *GRPCMetricsClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
