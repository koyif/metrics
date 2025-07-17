package main

import (
	"fmt"
	"github.com/koyif/metrics/internal/agent/app"
	"github.com/koyif/metrics/internal/agent/config"
	"log/slog"
	"os"
)

func main() {
	cfg := config.Load()
	a := app.New(cfg)

	if err := a.Run(); err != nil {
		slog.Error(fmt.Sprintf("error starting agent %v", err))
		os.Exit(1)
	}
}
