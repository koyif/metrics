package config

import (
	"flag"
	"os"
	"strconv"
	"time"
)

type ServerConfig struct {
	Addr string `yaml:"address" default:"localhost:8080"`
}

type StorageConfig struct {
	StoreInterval   time.Duration `yaml:"store_interval" default:"300s"`
	FileStoragePath string        `yaml:"file_storage_path" default:"/tmp/storage"`
	Restore         bool          `yaml:"restore" default:"false"`
}

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Storage StorageConfig `yaml:"storage"`
}

func Load() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Addr: "localhost:8080",
		},
		Storage: StorageConfig{
			StoreInterval:   300 * time.Second,
			FileStoragePath: "/tmp/storage",
			Restore:         false,
		},
	}

	flag.StringVar(&cfg.Server.Addr, "a", "localhost:8080", "адрес эндпоинта HTTP-сервера")
	flag.Func("i", "интервал сохранения", secondsToDuration(&cfg.Storage.StoreInterval))
	flag.StringVar(&cfg.Storage.FileStoragePath, "f", "/tmp/storage", "путь к файлу для хранения")
	flag.BoolVar(&cfg.Storage.Restore, "r", false, "восстанавливать данные из файла")

	flag.Parse()

	addressEnv := os.Getenv("ADDRESS")
	if addressEnv != "" {
		cfg.Server.Addr = addressEnv
	}

	return cfg
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
