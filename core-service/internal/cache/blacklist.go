package cache

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// TokenBlacklist marks JWTs as revoked (logout) until their natural
// expiry, backed by Redis so entries self-clean via TTL.
type TokenBlacklist struct {
	client *redis.Client
}

func NewTokenBlacklist(client *redis.Client) *TokenBlacklist {
	return &TokenBlacklist{client: client}
}

// Add revokes token for the remainder of its lifetime. No-op if ttl has
// already elapsed (token would be rejected on expiry anyway).
func (b *TokenBlacklist) Add(ctx context.Context, token string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	return b.client.Set(ctx, "auth:blacklist:"+token, "1", ttl).Err()
}

func (b *TokenBlacklist) Contains(ctx context.Context, token string) (bool, error) {
	n, err := b.client.Exists(ctx, "auth:blacklist:"+token).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
