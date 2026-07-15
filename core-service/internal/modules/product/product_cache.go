package product

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// Cache adalah kontrak minimal buat caching. Dibungkus interface (sama
// pola-nya kayak Repository) supaya service bisa di-unit-test tanpa Redis
// beneran — cukup pakai in-memory fake di test.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
}

// redisCache adalah adapter tipis di atas *redis.Client. errCacheMiss
// dikembalikan kalau key gak ketemu, biar service gak perlu tau detail
// redis.Nil dari library go-redis.
type redisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) Cache {
	return &redisCache{client: client}
}

var errCacheMiss = errCacheMissType{}

type errCacheMissType struct{}

func (errCacheMissType) Error() string { return "cache miss" }

func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", errCacheMiss
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (c *redisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}
