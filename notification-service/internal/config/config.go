package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	App     AppConfig
	Mailjet MailjetConfig
}

type AppConfig struct {
	Env  string
	Host string
	Port int
}

type MailjetConfig struct {
	APIKey    string
	APISecret string
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		log.Println("There's no .env file, using default values")
	}

	v.SetDefault("NOTIFICATION_ENV", "development")
	v.SetDefault("NOTIFICATION_HOST", "0.0.0.0")
	v.SetDefault("NOTIFICATION_PORT", 9002)

	cfg := &Config{
		App: AppConfig{
			Env:  v.GetString("NOTIFICATION_ENV"),
			Host: v.GetString("NOTIFICATION_HOST"),
			Port: v.GetInt("NOTIFICATION_PORT"),
		},
		Mailjet: MailjetConfig{
			APIKey:    v.GetString("MAILJET_API_KEY"),
			APISecret: v.GetString("MAILJET_API_SECRET"),
		},
	}

	return cfg, nil
}
