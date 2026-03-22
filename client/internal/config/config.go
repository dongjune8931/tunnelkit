package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	ServerURL  string `mapstructure:"server_url"`
	AuthToken  string `mapstructure:"auth_token"`
	Subdomain  string `mapstructure:"subdomain"`
	LocalPort  int    `mapstructure:"local_port"`
}

func Load() (*Config, error) {
	viper.SetDefault("server_url", "ws://localhost:8080")
	viper.SetDefault("local_port", 3000)

	viper.SetEnvPrefix("PREVIEWD")
	viper.AutomaticEnv()

	viper.SetConfigName(".previewd")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")
	_ = viper.ReadInConfig()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
