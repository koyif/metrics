package config

import "time"

type ServerConfig struct {
	Addr string `yaml:"addr"`
}

type Config struct {
	Server         ServerConfig  `yaml:"server"`
	PollInterval   time.Duration `yaml:"pollInterval"`
	ReportInterval time.Duration `yaml:"reportInterval"`
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Addr: "http://localhost:8080",
		},
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}
}
