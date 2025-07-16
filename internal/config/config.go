package config

import "flag"

type ServerConfig struct {
	Addr string `yaml:"address" default:"localhost:8080"`
}

type Config struct {
	Server ServerConfig `yaml:"server"`
}

func Load() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Addr: "localhost:8080",
		},
	}

	flag.StringVar(&cfg.Server.Addr, "a", "localhost:8080", "адрес эндпоинта HTTP-сервера")

	flag.Parse()

	return cfg
}
