package config

import "flag"

type ServerConfig struct {
	Addr string `yaml:"addr"`
}

type Config struct {
	Server         ServerConfig `yaml:"server"`
	PollInterval   int          `yaml:"pollInterval"`
	ReportInterval int          `yaml:"reportInterval"`
}

func Load() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Addr: "localhost:8080",
		},
		PollInterval:   2,
		ReportInterval: 10,
	}

	flag.IntVar(&cfg.PollInterval, "p", 2, "частота опроса метрик из пакета runtime в секундах")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "частота отправки метрик на сервер в секундах")
	flag.StringVar(&cfg.Server.Addr, "a", "localhost:8080", "адрес эндпоинта HTTP-сервера")

	flag.Parse()

	return cfg
}
