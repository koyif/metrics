package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"

	"github.com/koyif/metrics/pkg/types"
)

type ServerConfig struct {
}

type Config struct {
	Addr           string                  `json:"address" env:"ADDRESS" env-default:"localhost:8080"`
	PollInterval   types.DurationInSeconds `json:"poll_interval" env:"POLL_INTERVAL" env-default:"2"`
	ReportInterval types.DurationInSeconds `json:"report_interval" env:"REPORT_INTERVAL" env-default:"10"`
	HashKey        string                  `json:"hash_key" env:"KEY"`
	RateLimit      int                     `json:"rate_limit" env:"RATE_LIMIT" env-default:"3"`
	CryptoKey      string                  `json:"crypto_key" env:"CRYPTO_KEY"`
	UseGRPC        bool                    `json:"use_grpc" env:"USE_GRPC" env-default:"false"`
	ConfigPath     string                  `json:"-"`
}

func Load() (*Config, error) {
	cfg := &Config{}

	setupFlags(cfg)

	if cfg.ConfigPath == "" {
		cfg.ConfigPath = os.Getenv("CONFIG")
	}

	if cfg.ConfigPath != "" {
		loadJSONConfig(cfg)
	}

	applyEnvVars(cfg)

	setupFlags(cfg)

	log.Printf("loaded config: %+v", cfg)

	return cfg, nil
}

func loadJSONConfig(cfg *Config) {
	err := cleanenv.ReadConfig(cfg.ConfigPath, cfg)
	if err != nil {
		log.Fatalf("couldn't read JSON config from %s: %s", cfg.ConfigPath, err)
		return
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

	flag.Func("p", "частота опроса метрик из пакета runtime в секундах", func(s string) error { return cfg.PollInterval.SetValue(s) })
	flag.Func("r", "частота отправки метрик на сервер в секундах", func(s string) error { return cfg.ReportInterval.SetValue(s) })
	flag.StringVar(&cfg.HashKey, "k", cfg.HashKey, "ключ для хеширования")
	flag.IntVar(&cfg.RateLimit, "l", cfg.RateLimit, "лимит одновременной отправки метрик")
	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "адрес эндпоинта HTTP-сервера")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "путь до файла с публичным ключом")
	flag.BoolVar(&cfg.UseGRPC, "use-grpc", cfg.UseGRPC, "использовать gRPC вместо HTTP")

	flag.Parse()
}
