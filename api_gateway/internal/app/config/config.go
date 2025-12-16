package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	Env      string        `env:"ENV" env-default:"local"`
	HTTP     HTTPConfig    `env-prefix:"HTTP_"`
	Storage  StorageConfig `env-prefix:"STORAGE_"`
	Analysis AnalysisConfig `env-prefix:"ANALYSIS_"`
}

type HTTPConfig struct {
	Port int `env:"PORT" env-default:"8080"`
}

type StorageConfig struct {
	Addr string `env:"ADDR" env-default:"storage-service:5001"`
}

type AnalysisConfig struct {
	Addr string `env:"ADDR" env-default:"plagiarism-service:6001"`
}

func MustLoad() *Config {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}
	return &cfg
}


