package main

import (
	"log"

	"github.com/koyif/metrics/internal/agent/app"
	"github.com/koyif/metrics/internal/agent/config"
	"github.com/koyif/metrics/internal/app/logger"
)

func main() {
	cfg := config.Load()
	if err := logger.Initialize(); err != nil {
		log.Fatalf("error starting logger: %v", err)
	}

	a := app.New(cfg)

	if err := a.Run(); err != nil {
		logger.Log.Fatal("error running agent", logger.Error(err))
	}
}
