package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/koyif/metrics/internal/app"
	"github.com/koyif/metrics/internal/app/logger"
	"github.com/koyif/metrics/internal/config"
)

func main() {
	cfg := config.Load()
	if err := logger.Initialize(); err != nil {
		logger.Log.Fatal("error starting logger", logger.Error(err))
	}

	application := app.New(cfg)

	if err := run(application); err != nil {
		logger.Log.Fatal("error starting server", logger.Error(err))
	}
}

func run(a *app.App) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		logger.Log.Info("starting server", logger.String("address", a.Config.Server.Addr))
		err := http.ListenAndServe(a.Config.Server.Addr, a.Router())
		if err != nil {
			logger.Log.Fatal("error starting server", logger.Error(err))
			stop <- syscall.SIGTERM
		}
	}()

	<-stop
	logger.Log.Info("shutting down")
	err := a.MetricsService.Persist()
	if err != nil {
		return err
	}

	return nil
}
