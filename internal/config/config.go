package config

import (
	"flag"
	"github.com/koyif/metrics/internal/app/logger"
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

	getEnv(cfg)

	return cfg
}

func getEnv(cfg *Config) {

	addressEnv := os.Getenv("ADDRESS")
	if addressEnv != "" {
		cfg.Server.Addr = addressEnv
	}

	restoreEnv := os.Getenv("RESTORE")
	if restoreEnv != "" {
		restore, err := strconv.ParseBool(restoreEnv)
		if err != nil {
			logger.Log.Error("couldn't get environment variable", logger.String("env variable", "RESTORE"))
		} else {
			cfg.Storage.Restore = restore
		}
	}

	storeIntervalEnv := os.Getenv("STORE_INTERVAL")
	if storeIntervalEnv != "" {
		storeInterval, err := strconv.ParseInt(storeIntervalEnv, 10, 64)
		if err != nil {
			logger.Log.Error("couldn't get environment variable", logger.String("env variable", "STORE_INTERVAL"))
		} else {
			cfg.Storage.StoreInterval = time.Duration(storeInterval) * time.Second
		}
	}

	fileStoragePathEnv := os.Getenv("FILE_STORAGE_PATH")
	if fileStoragePathEnv != "" {
		cfg.Storage.FileStoragePath = fileStoragePathEnv
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
