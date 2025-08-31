package config

import (
	"flag"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/koyif/metrics/pkg/types"
)

const (
	AddressEnvVarName        = "ADDRESS"
	ReportIntervalEnvVarName = "REPORT_INTERVAL"
	PollIntervalEnvVarName   = "POLL_INTERVAL"
	HashKeyEnvVarName        = "KEY"
)

type ServerConfig struct {
	Addr string `yaml:"addr" env:"ADDRESS" env-default:"localhost:8080"`
}

type Config struct {
	Server         ServerConfig            `yaml:"server"`
	PollInterval   types.DurationInSeconds `yaml:"pollInterval" env:"POLL_INTERVAL" env-default:"2"`
	ReportInterval types.DurationInSeconds `yaml:"reportInterval" env:"REPORT_INTERVAL" env-default:"10"`
	HashKey        string                  `yaml:"hashKey" env:"KEY"`
}

func Load() *Config {
	cfg := &Config{}

	flag.Func("p", "частота опроса метрик из пакета runtime в секундах", func(s string) error { return cfg.PollInterval.SetValue(s) })
	flag.Func("r", "частота отправки метрик на сервер в секундах", func(s string) error { return cfg.ReportInterval.SetValue(s) })
	flag.StringVar(&cfg.HashKey, "k", "", "ключ для хеширования")
	flag.StringVar(&cfg.Server.Addr, "a", "localhost:8080", "адрес эндпоинта HTTP-сервера")

	flag.Parse()

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		log.Fatalf("couldn't read environment variables: %s", err.Error())
	}

	log.Printf("loaded config: %+v", cfg)

	return cfg
}
