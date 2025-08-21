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
	DatabaseURL     string        `yaml:"database_url"`
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
	flag.StringVar(&cfg.Storage.DatabaseURL, "d", "", "URL базы данных для хранения метрик")

	flag.Parse()

	getEnvString("ADDRESS", &cfg.Server.Addr)
	getEnvInt("STORE_INTERVAL", &cfg.Storage.StoreInterval, time.Second)
	getEnvBool("RESTORE", &cfg.Storage.Restore)
	getEnvString("FILE_STORAGE_PATH", &cfg.Storage.FileStoragePath)
	getEnvString("DATABASE_CONN_STRING", &cfg.Storage.DatabaseURL)

	return cfg
}

func getEnvString(envKey string, target *string) {
	if value := os.Getenv(envKey); value != "" {
		*target = value
	}
}

func getEnvBool(envKey string, target *bool) {
	if value := os.Getenv(envKey); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			*target = boolVal
		}
	}
}

func getEnvInt(envKey string, target *time.Duration, multiplier time.Duration) {
	if value := os.Getenv(envKey); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			*target = time.Duration(intVal) * multiplier
		}
	}
}

func secondsToDuration(interval *time.Duration) func(string) error {
	return func(s string) error {
		if sec, err := strconv.Atoi(s); err != nil {
			return err
		} else if sec < 0 {
			return nil
		} else {
			*interval = time.Duration(sec) * time.Second
		}

		return nil
	}
}
