package interceptor

import (
	"context"
	"fmt"
	"net"

	"github.com/koyif/metrics/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// IPCheckInterceptor creates a gRPC UnaryServerInterceptor that validates client IP against trusted subnet.
// Returns a pass-through interceptor if trustedSubnet is empty (no validation).
func IPCheckInterceptor(trustedSubnet string) grpc.UnaryServerInterceptor {
	if trustedSubnet == "" {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}

	_, ipNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		logger.Log.Fatal("invalid CIDR notation for trusted subnet", logger.Error(err))
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return nil, status.Error(codes.Internal, "server misconfiguration")
		}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Log.Warn(
				"missing metadata in gRPC request",
				logger.String("method", info.FullMethod),
			)
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}

		ips := md.Get("x-real-ip")
		if len(ips) == 0 {
			logger.Log.Warn(
				"missing x-real-ip in gRPC metadata",
				logger.String("method", info.FullMethod),
			)
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}

		clientIP := ips[0]

		ip := net.ParseIP(clientIP)
		if ip == nil {
			logger.Log.Warn(
				"invalid IP address in x-real-ip metadata",
				logger.String("x-real-ip", clientIP),
				logger.String("method", info.FullMethod),
			)
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}

		if !ipNet.Contains(ip) {
			logger.Log.Warn(
				"IP address not in trusted subnet",
				logger.String("x-real-ip", clientIP),
				logger.String("trusted_subnet", trustedSubnet),
				logger.String("method", info.FullMethod),
			)
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}

		logger.Log.Debug(
			"IP check passed",
			logger.String("x-real-ip", clientIP),
			logger.String("trusted_subnet", trustedSubnet),
			logger.String("method", info.FullMethod),
		)

		return handler(ctx, req)
	}
}

// ExtractIPFromMetadata extracts the client IP from gRPC metadata.
// Returns the IP address and an error if not found or invalid.
func ExtractIPFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("missing metadata")
	}

	ips := md.Get("x-real-ip")
	if len(ips) == 0 {
		return "", fmt.Errorf("missing x-real-ip in metadata")
	}

	return ips[0], nil
}
