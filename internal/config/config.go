package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"

	"github.com/koyif/metrics/pkg/types"
)

type Config struct {
	Addr            string                  `json:"address" env:"ADDRESS" env-default:"localhost:8080"`
	StoreInterval   types.DurationInSeconds `json:"store_interval" env:"STORE_INTERVAL" env-default:"300"`
	FileStoragePath string                  `json:"store_file" env:"FILE_STORAGE_PATH" env-default:"/tmp/storage"`
	Restore         bool                    `json:"restore" env:"RESTORE" env-default:"false"`
	DatabaseURL     string                  `json:"database_dsn" env:"DATABASE_DSN"`
	FilePath        string                  `json:"audit_file" env:"AUDIT_FILE"`
	URL             string                  `json:"audit_url" env:"AUDIT_URL"`
	HashKey         string                  `json:"hash_key" env:"KEY"`
	CryptoKey       string                  `json:"crypto_key" env:"CRYPTO_KEY"`
	TrustedSubnet   string                  `json:"trusted_subnet" env:"TRUSTED_SUBNET"`
	GRPCAddr        string                  `json:"grpc_address" env:"GRPC_ADDRESS"`
	ConfigPath      string                  `json:"-"`
}

func Load() *Config {
	cfg := &Config{}

	setupFlags(cfg)

	if cfg.ConfigPath == "" {
		cfg.ConfigPath = os.Getenv("CONFIG")
	}

	// Priority (high -> low):
	// 1. Flags
	// 2. Environment
	// 3. File

	if cfg.ConfigPath != "" {
		loadJSONConfig(cfg)
	}

	applyEnvVars(cfg)

	setupFlags(cfg)

	log.Printf("loaded config: %+v", cfg)

	return cfg
}

func loadJSONConfig(cfg *Config) {
	err := cleanenv.ReadConfig(cfg.ConfigPath, cfg)
	if err != nil {
		log.Fatalf("couldn't read JSON config from %s: %s", cfg.ConfigPath, err.Error())
	}

	log.Printf("loaded JSON config from %s", cfg.ConfigPath)
}

func applyEnvVars(cfg *Config) {
	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		log.Fatalf("couldn't read environment variables: %s", err.Error())
	}
}

func setupFlags(cfg *Config) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	flag.StringVar(&cfg.ConfigPath, "c", cfg.ConfigPath, "путь к файлу конфигурации")
	flag.StringVar(&cfg.ConfigPath, "config", cfg.ConfigPath, "путь к файлу конфигурации")

	flag.Func("i", "интервал сохранения", func(s string) error { return cfg.StoreInterval.SetValue(s) })
	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "адрес эндпоинта HTTP-сервера")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "путь к файлу для хранения")
	flag.BoolVar(&cfg.Restore, "r", cfg.Restore, "восстанавливать данные из файла")
	flag.StringVar(&cfg.DatabaseURL, "d", cfg.DatabaseURL, "URL базы данных для хранения метрик")
	flag.StringVar(&cfg.HashKey, "k", cfg.HashKey, "ключ для хеширования")
	flag.StringVar(&cfg.FilePath, "audit-file", cfg.FilePath, "путь к файлу для логов аудита")
	flag.StringVar(&cfg.URL, "audit-url", cfg.URL, "URL для отправки логов аудита")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "путь до файла с приватным ключом")
	flag.StringVar(&cfg.TrustedSubnet, "t", cfg.TrustedSubnet, "доверенная подсеть в формате CIDR")
	flag.StringVar(&cfg.GRPCAddr, "g", cfg.GRPCAddr, "адрес gRPC-сервера")

	// Parse flags again - command-line flags will override JSON/env values
	flag.Parse()
}
