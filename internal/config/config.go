package config

import (
	"flag"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/koyif/metrics/pkg/types"
)

type ServerConfig struct {
	Addr string `yaml:"address" env:"ADDRESS" env-default:"localhost:8080"`
}

type StorageConfig struct {
	StoreInterval   types.DurationInSeconds `yaml:"store_interval" env:"STORE_INTERVAL" env-default:"300"`
	FileStoragePath string                  `yaml:"file_storage_path" env:"FILE_STORAGE_PATH" env-default:"/tmp/storage"`
	Restore         bool                    `yaml:"restore" env:"RESTORE" env-default:"false"`
	DatabaseURL     string                  `yaml:"database_url" env:"DATABASE_DSN"`
}

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Storage StorageConfig `yaml:"storage"`
	HashKey string        `yaml:"hashKey" env:"KEY"`
}

func Load() *Config {
	cfg := &Config{}

	flag.Func("i", "интервал сохранения", func(s string) error { return cfg.Storage.StoreInterval.SetValue(s) })
	flag.StringVar(&cfg.Server.Addr, "a", "localhost:8080", "адрес эндпоинта HTTP-сервера")
	flag.StringVar(&cfg.Storage.FileStoragePath, "f", "/tmp/storage", "путь к файлу для хранения")
	flag.BoolVar(&cfg.Storage.Restore, "r", false, "восстанавливать данные из файла")
	flag.StringVar(&cfg.Storage.DatabaseURL, "d", "", "URL базы данных для хранения метрик")
	flag.StringVar(&cfg.HashKey, "k", "", "ключ для хеширования")

	flag.Parse()

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		log.Fatalf("couldn't read environment variables: %s", err.Error())
	}

	log.Printf("loaded config: %+v", cfg)

	return cfg
}
