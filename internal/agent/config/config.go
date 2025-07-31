package config

import (
	"flag"
	"log/slog"
	"os"
	"strconv"
	"time"
)

const (
	AddressEnvVarName        = "ADDRESS"
	ReportIntervalEnvVarName = "REPORT_INTERVAL"
	PollIntervalEnvVarName   = "POLL_INTERVAL"
)

type ServerConfig struct {
	Addr string `yaml:"addr"`
}

type Config struct {
	Server         ServerConfig  `yaml:"server"`
	PollInterval   time.Duration `yaml:"pollInterval"`
	ReportInterval time.Duration `yaml:"reportInterval"`
}

func Load() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Addr: "localhost:8080",
		},
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}

	flag.Func("p", "частота опроса метрик из пакета runtime в секундах", secondsToDuration(&cfg.PollInterval))
	flag.Func("r", "частота отправки метрик на сервер в секундах", secondsToDuration(&cfg.ReportInterval))
	flag.StringVar(&cfg.Server.Addr, "a", "localhost:8080", "адрес эндпоинта HTTP-сервера")

	flag.Parse()

	loadEnv(cfg)

	return cfg
}

func loadEnv(cfg *Config) {
	addressEnv := os.Getenv(AddressEnvVarName)
	if addressEnv != "" {
		cfg.Server.Addr = addressEnv
	}
	reportIntervalEnv := os.Getenv(ReportIntervalEnvVarName)
	if reportIntervalEnv != "" {
		err := secondsToDuration(&cfg.ReportInterval)(reportIntervalEnv)
		if err != nil {
			slog.Error("couldn't get environment variable", slog.StringValue(ReportIntervalEnvVarName))
		}
	}
	pollIntervalEnv := os.Getenv(PollIntervalEnvVarName)
	if pollIntervalEnv != "" {
		err := secondsToDuration(&cfg.PollInterval)(pollIntervalEnv)
		if err != nil {
			slog.Error("couldn't get environment variable", slog.StringValue(PollIntervalEnvVarName))
		}
	}
}

func secondsToDuration(interval *time.Duration) func(string) error {
	return func(s string) error {
		sec, err := strconv.Atoi(s)
		if err != nil {
			return err
		}

		*interval = time.Duration(sec) * time.Second

		return nil
	}
}
