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

type AuditConfig struct {
	FilePath string `yaml:"audit_file" env:"AUDIT_FILE"`
	URL      string `yaml:"audit_url" env:"AUDIT_URL"`
}

type StorageConfig struct {
	StoreInterval   types.DurationInSeconds `yaml:"store_interval" env:"STORE_INTERVAL" env-default:"300"`
	FileStoragePath string                  `yaml:"file_storage_path" env:"FILE_STORAGE_PATH" env-default:"/tmp/storage"`
	Restore         bool                    `yaml:"restore" env:"RESTORE" env-default:"false"`
	DatabaseURL     string                  `yaml:"database_url" env:"DATABASE_DSN"`
}

type Config struct {
	Server    ServerConfig  `yaml:"server"`
	Storage   StorageConfig `yaml:"storage"`
	Audit     AuditConfig   `yaml:"audit"`
	HashKey   string        `yaml:"hashKey" env:"KEY"`
	CryptoKey string        `yaml:"cryptoKey" env:"CRYPTO_KEY"`
}

func Load() *Config {
	cfg := &Config{}

	flag.Func("i", "интервал сохранения", func(s string) error { return cfg.Storage.StoreInterval.SetValue(s) })
	flag.StringVar(&cfg.Server.Addr, "a", "localhost:8080", "адрес эндпоинта HTTP-сервера")
	flag.StringVar(&cfg.Storage.FileStoragePath, "f", "/tmp/storage", "путь к файлу для хранения")
	flag.BoolVar(&cfg.Storage.Restore, "r", false, "восстанавливать данные из файла")
	flag.StringVar(&cfg.Storage.DatabaseURL, "d", "", "URL базы данных для хранения метрик")
	flag.StringVar(&cfg.HashKey, "k", "", "ключ для хеширования")
	flag.StringVar(&cfg.Audit.FilePath, "audit-file", "", "путь к файлу для логов аудита")
	flag.StringVar(&cfg.Audit.URL, "audit-url", "", "URL для отправки логов аудита")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "путь до файла с приватным ключом")

	flag.Parse()

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		log.Fatalf("couldn't read environment variables: %s", err.Error())
	}

	log.Printf("loaded config: %+v", cfg)

	return cfg
}
