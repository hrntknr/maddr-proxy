package main

import (
	"fmt"
	"strings"

	maddrproxy "github.com/hrntknr/maddr-proxy/pkg/maddr-proxy"
	"github.com/spf13/viper"
)

type Config struct {
	Listen   string
	Password string
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	var config Config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/maddr-proxy/")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetDefault("listen", ":1080")
	viper.ReadInConfig()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config: %w", err)
		}
	}
	if err := viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	passwords := strings.Split(config.Password, ",")
	if len(passwords) == 1 && passwords[0] == "" {
		passwords = []string{}
	}
	if err := maddrproxy.NewProxy(passwords).ListenAndServe(config.Listen); err != nil {
		return fmt.Errorf("failed to start proxy: %w", err)
	}
	return nil
}
