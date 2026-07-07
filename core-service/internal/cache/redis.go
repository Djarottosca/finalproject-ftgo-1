package cache

import (
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/config"
)

func NewRedis(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return client, nil
}
