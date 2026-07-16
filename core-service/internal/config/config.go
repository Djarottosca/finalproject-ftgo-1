package config

import (
	"fmt"
	"log"
	"net/url"

	"github.com/spf13/viper"
)

type Config struct {
	App           AppConfig
	Database      DatabaseConfig
	Redis         RedisConfig
	Notification  NotificationConfig
	JWTSecret     string

	// PaymentServiceAddr is the gRPC address of payment-service. It differs
	// per environment (localhost in dev, a service name in deploy), so it's
	// config, never hardcoded.
	PaymentServiceAddr string
}

// NotificationConfig: alamat gRPC notification-service, dibaca dari
// NOTIFICATION_HOST/NOTIFICATION_PORT di .env (bukan NOTIFICATION_ENV/PORT
// milik service itu sendiri, ini alamat buat core konek ke sana).
type NotificationConfig struct {
	Host string
	Port int
}

func (c *NotificationConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type AppConfig struct {
	Env  string
	Host string
	Port int
}

type DatabaseConfig struct {
	Host    string
	Port    int
	User    string
	Pass    string
	Name    string
	SSLMode string
}

func (c *DatabaseConfig) DSN() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Pass),
		Host:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:     c.Name,
		RawQuery: fmt.Sprintf("sslmode=%s", c.SSLMode),
	}

	return u.String()
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		log.Println("There's no .env file, using default values")
	}

	v.SetDefault("CORE_ENV", "development")
	v.SetDefault("CORE_HOST", "0.0.0.0")
	v.SetDefault("CORE_PORT", 8080)
	v.SetDefault("DB_PORT", 5432)
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("REDIS_PORT", 6379)
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("PAYMENT_SERVICE_ADDR", "localhost:9001")
	v.SetDefault("NOTIFICATION_HOST", "localhost")
	v.SetDefault("NOTIFICATION_PORT", 9002)

	cfg := &Config{
		App: AppConfig{
			Env:  v.GetString("CORE_ENV"),
			Host: v.GetString("CORE_HOST"),
			Port: v.GetInt("CORE_PORT"),
		},
		Database: DatabaseConfig{
			Host:    v.GetString("DB_HOST"),
			Port:    v.GetInt("DB_PORT"),
			User:    v.GetString("DB_USER"),
			Pass:    v.GetString("DB_PASSWORD"),
			Name:    v.GetString("DB_NAME"),
			SSLMode: v.GetString("DB_SSLMODE"),
		},
		Redis: RedisConfig{
			Host:     v.GetString("REDIS_HOST"),
			Port:     v.GetInt("REDIS_PORT"),
			Password: v.GetString("REDIS_PASSWORD"),
			DB:       v.GetInt("REDIS_DB"),
		},
		PaymentServiceAddr: v.GetString("PAYMENT_SERVICE_ADDR"),
		Notification: NotificationConfig{
			Host: v.GetString("NOTIFICATION_HOST"),
			Port: v.GetInt("NOTIFICATION_PORT"),
		},
		JWTSecret: v.GetString("JWT_SECRET"),
	}

	return cfg, nil
}
