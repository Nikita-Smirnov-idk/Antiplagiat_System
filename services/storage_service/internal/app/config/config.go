package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	StoragePath string         `env:"STORAGE_PATH" env-required:"true"`
	GRPC        GRPCConfig     `env-prefix:"GRPC_"`
	DB          PostgresConfig `env-prefix:"POSTGRES_"`
}

type PostgresConfig struct {
	Path    string `env:"PATH" env-required:"true"`
	MaxConn int    `env:"MAX_CONN" env-default:"20"`
	MinConn int    `env:"MIN_CONN" env-default:"5"`
}

type GRPCConfig struct {
	Port    int           `env:"PORT" env-default:"5001"`
	TimeOut time.Duration `env:"TIMEOUT" env-default:"5s"`
}

func MustLoad() *Config {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}
