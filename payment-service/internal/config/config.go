package config

import (
	"log"
	"time"

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
	BaseURL  string
}

type ProviderConfig struct {
	Name       string        // "simulation" | "xendit"
	InvoiceTTL time.Duration // umur invoice simulasi
}

type XenditConfig struct {
	APIKey      string
	BaseURL     string
	ReturnURL   string
	HTTPTimeout time.Duration // timeout request ke Xendit
	Currency    string        // "IDR"
	Country     string        // "ID"
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
	v.SetDefault("PAYMENT_BASE_URL", "http://localhost:9001")
	v.SetDefault("PAYMENT_PROVIDER", "simulation")
	v.SetDefault("XENDIT_BASE_URL", "https://api.xendit.co")
	v.SetDefault("XENDIT_RETURN_URL", "https://example.com/orders")
	v.SetDefault("PAYMENT_INVOICE_TTL", "24h")
	v.SetDefault("XENDIT_HTTP_TIMEOUT", "15s")
	v.SetDefault("XENDIT_CURRENCY", "IDR")
	v.SetDefault("XENDIT_COUNTRY", "ID")

	cfg := &Config{
		App: AppConfig{
			Env:      v.GetString("PAYMENT_ENV"),
			Host:     v.GetString("PAYMENT_HOST"),
			GRPCPort: v.GetInt("PAYMENT_GRPC_PORT"),
			BaseURL:  v.GetString("PAYMENT_BASE_URL"),
		},
		Provider: ProviderConfig{
			Name:       v.GetString("PAYMENT_PROVIDER"),
			InvoiceTTL: v.GetDuration("PAYMENT_INVOICE_TTL"),
		},
		Xendit: XenditConfig{
			APIKey:      v.GetString("XENDIT_API_KEY"),
			BaseURL:     v.GetString("XENDIT_BASE_URL"),
			ReturnURL:   v.GetString("XENDIT_RETURN_URL"),
			HTTPTimeout: v.GetDuration("XENDIT_HTTP_TIMEOUT"),
			Currency:    v.GetString("XENDIT_CURRENCY"),
			Country:     v.GetString("XENDIT_COUNTRY"),
		},
	}

	return cfg, nil
}
