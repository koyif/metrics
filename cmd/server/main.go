package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/koyif/metrics/internal/app"
	"github.com/koyif/metrics/internal/config"
	grpcinterceptor "github.com/koyif/metrics/internal/grpc/interceptor"
	grpcserver "github.com/koyif/metrics/internal/grpc/server"
	"github.com/koyif/metrics/internal/proto/api/proto"
	"github.com/koyif/metrics/pkg/logger"
	"google.golang.org/grpc"

	_ "github.com/koyif/metrics/docs"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

//	@title			Metrics Collection API
//	@version		1.0
//	@description	A metrics collection and alerting server for storing and retrieving counter and gauge metrics
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.email	support@example.com

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

//	@host		localhost:8080
//	@BasePath	/

//	@schemes	http https

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

func main() {
	printBuildInfo()

	cfg := config.Load()
	if err := logger.Initialize(); err != nil {
		log.Fatalf("error starting logger: %v", err)
	}

	runMigrations(cfg)

	wg := sync.WaitGroup{}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	if cfg.StoreInterval.Value() != 0 {
		wg.Add(1)
	}

	application, err := app.New(ctx, &wg, cfg)
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}

	server := startServer(application)

	var grpcSrv *grpc.Server
	if cfg.GRPCAddr != "" {
		grpcSrv = startGRPCServer(application, &wg)
	}

	<-ctx.Done()
	logger.Log.Info("shutting down gracefully")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown HTTP server gracefully
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("server shutdown error", logger.Error(err))
	}

	// Shutdown gRPC server gracefully
	if grpcSrv != nil {
		logger.Log.Info("shutting down gRPC server")
		grpcSrv.GracefulStop()
	}

	// Wait for background tasks (file persistence) to complete
	wg.Wait()
	logger.Log.Info("shutdown complete")
}

func runMigrations(cfg *config.Config) {
	logger.Log.Info("running database migrations")
	app.RunMigrations(cfg.DatabaseURL)
}

func startServer(a *app.App) *http.Server {
	server := &http.Server{
		Addr:    a.Config.Addr,
		Handler: a.Router(),
	}

	go func() {
		logger.Log.Info("starting server", logger.String("address", a.Config.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Error("server error", logger.Error(err))
		}
	}()

	return server
}

func startGRPCServer(a *app.App, wg *sync.WaitGroup) *grpc.Server {
	wg.Add(1)

	interceptor := grpcinterceptor.IPCheckInterceptor(a.Config.TrustedSubnet)
	grpcSrv := grpc.NewServer(grpc.UnaryInterceptor(interceptor))
	metricsServer := grpcserver.NewMetricsServer(a.MetricsService, a.Config, a.AuditManager)
	proto.RegisterMetricsServer(grpcSrv, metricsServer)

	lis, err := net.Listen("tcp", a.Config.GRPCAddr)
	if err != nil {
		logger.Log.Fatal("failed to listen for gRPC", logger.Error(err))
	}

	go func() {
		defer wg.Done()
		logger.Log.Info("starting gRPC server", logger.String("address", a.Config.GRPCAddr))
		if err := grpcSrv.Serve(lis); err != nil {
			logger.Log.Error("gRPC server error", logger.Error(err))
		}
	}()

	return grpcSrv
}
