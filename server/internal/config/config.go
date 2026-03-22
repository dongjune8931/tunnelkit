package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	ServerHost   string `mapstructure:"server_host"`
	TunnelPort   int    `mapstructure:"tunnel_port"`
	DashboardPort int   `mapstructure:"dashboard_port"`
	DBPath       string `mapstructure:"db_path"`
	AdminToken   string `mapstructure:"admin_token"`
}

func Load() (*Config, error) {
	viper.SetDefault("server_host", "localhost")
	viper.SetDefault("tunnel_port", 8080)
	viper.SetDefault("dashboard_port", 8081)
	viper.SetDefault("db_path", "./previewd.db")
	viper.SetDefault("admin_token", "changeme")

	viper.SetEnvPrefix("PREVIEWD")
	viper.AutomaticEnv()

	viper.SetConfigName("previewd")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	_ = viper.ReadInConfig()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
