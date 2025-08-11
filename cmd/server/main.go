package main

import (
	"context"
	"errors"
	"github.com/koyif/metrics/internal/app"
	"github.com/koyif/metrics/internal/app/logger"
	"github.com/koyif/metrics/internal/config"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	cfg := config.Load()
	if err := logger.Initialize(); err != nil {
		logger.Log.Fatal("error starting logger", logger.Error(err))
	}

	wg := sync.WaitGroup{}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if cfg.Storage.StoreInterval != 0 {
		wg.Add(1)
	}

	application := app.New(ctx, &wg, cfg)

	go startServer(application)

	<-ctx.Done()
	logger.Log.Info("shutting down")
	wg.Wait()
}

func startServer(a *app.App) {
	logger.Log.Info("starting server", logger.String("address", a.Config.Server.Addr))
	if err := http.ListenAndServe(a.Config.Server.Addr, a.Router()); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Log.Error("server error", logger.Error(err))
	}
}
