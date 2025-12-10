package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"sync"
	"syscall"

	"github.com/koyif/metrics/internal/app"
	"github.com/koyif/metrics/internal/config"
	"github.com/koyif/metrics/pkg/logger"

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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if cfg.StoreInterval.Value() != 0 {
		wg.Add(1)
	}

	application, err := app.New(ctx, &wg, cfg)
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}

	go startServer(application)

	<-ctx.Done()
	logger.Log.Info("shutting down")
	wg.Wait()
}

func runMigrations(cfg *config.Config) {
	logger.Log.Info("running database migrations")
	app.RunMigrations(cfg.DatabaseURL)
}

func startServer(a *app.App) {
	logger.Log.Info("starting server", logger.String("address", a.Config.Addr))
	if err := http.ListenAndServe(a.Config.Addr, a.Router()); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Log.Error("server error", logger.Error(err))
	}
}
