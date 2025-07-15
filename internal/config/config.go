package config

type ServerConfig struct {
	Port int `yaml:"port" default:"8080"`
}

type Config struct {
	Server ServerConfig `yaml:"server"`
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8080,
		},
	}
}
