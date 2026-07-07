package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	App    AppConfig
	Xendit XenditConfig
}

type AppConfig struct {
	Env  string
	Host string
	Port int
}

type XenditConfig struct {
	APIKey string
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		log.Println("There's no .env file, using default values")
	}

	v.SetDefault("PAYMENT_ENV", "development")
	v.SetDefault("PAYMENT_HOST", "0.0.0.0")
	v.SetDefault("PAYMENT_PORT", 9001)

	cfg := &Config{
		App: AppConfig{
			Env:  v.GetString("PAYMENT_ENV"),
			Host: v.GetString("PAYMENT_HOST"),
			Port: v.GetInt("PAYMENT_PORT"),
		},
		Xendit: XenditConfig{
			APIKey: v.GetString("XENDIT_API_KEY"),
		},
	}

	return cfg, nil
}
