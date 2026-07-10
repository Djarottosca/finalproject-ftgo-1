package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Provider ProviderConfig
	Xendit   XenditConfig
}

type AppConfig struct {
	Env      string
	Host     string
	GRPCPort int
	HTTPPort int
	BaseURL  string // URL publik payment-service, buat nyusun link simulasi
}

type ProviderConfig struct {
	Name string // "simulation" | "xendit"
}

type XenditConfig struct {
	APIKey        string
	CallbackToken string
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
	v.SetDefault("PAYMENT_GRPC_PORT", 9001)
	v.SetDefault("PAYMENT_HTTP_PORT", 9002)
	v.SetDefault("PAYMENT_BASE_URL", "http://localhost:9002")
	v.SetDefault("PAYMENT_PROVIDER", "simulation")

	cfg := &Config{
		App: AppConfig{
			Env:      v.GetString("PAYMENT_ENV"),
			Host:     v.GetString("PAYMENT_HOST"),
			GRPCPort: v.GetInt("PAYMENT_GRPC_PORT"),
			HTTPPort: v.GetInt("PAYMENT_HTTP_PORT"),
			BaseURL:  v.GetString("PAYMENT_BASE_URL"),
		},
		Provider: ProviderConfig{
			Name: v.GetString("PAYMENT_PROVIDER"),
		},
		Xendit: XenditConfig{
			APIKey:        v.GetString("XENDIT_API_KEY"),
			CallbackToken: v.GetString("XENDIT_CALLBACK_TOKEN"),
		},
	}

	return cfg, nil
}
