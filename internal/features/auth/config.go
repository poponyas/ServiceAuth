package grpc_config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port int `envconfig:"PORT"`

	Timeout time.Duration `envconfig:"TIMEOUT"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("GRPC", &config); err != nil {
		return Config{}, fmt.Errorf("can't get grpc config vars:%w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("can't get grpc config")
		panic(err)
	}

	return config
}
