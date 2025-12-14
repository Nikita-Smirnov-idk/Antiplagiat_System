package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env  string         `env:"ENV" env-default:"local"`
	GRPC GRPCConfig     `env-prefix:"GRPC_"`
	DB   PostgresConfig `env-prefix:"POSTGRES_"`
	S3   S3Config       `env-prefix:"S3_"`
}

type PostgresConfig struct {
	Path        string `env:"PATH" env-required:"true"`
	MaxConn     int32  `env:"MAX_CONN" env-default:"20"`
	MinConn     int32  `env:"MIN_CONN" env-default:"5"`
	AutoMigrate bool   `env:"AUTO_MIGRATE" env-default:"true"`
}

type GRPCConfig struct {
	Port    int           `env:"PORT" env-default:"5001"`
	TimeOut time.Duration `env:"TIMEOUT" env-default:"5s"`
}

type S3Config struct {
	Bucket         string `env:"BUCKET" env-required:"true"`
	Region         string `env:"REGION" env-required:"true"`
	Endpoint       string `env:"ENDPOINT" env-required:"true"`
	ClientEndpoint string `env:"CLIENT_ENDPOINT" env-required:"true"`
	AccessKey      string `env:"ACCESS_KEY" env-required:"true"`
	SecretKey      string `env:"SECRET_KEY" env-required:"true"`
	ExpirationTime int32  `env:"EXPIRATION_TIME" env-default:"5"`
}

func MustLoad() *Config {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}
