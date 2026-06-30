package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(addr, password string, db int, ttl time.Duration) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20,
		MinIdleConns: 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	log.Println("Connected to Redis")
	return &RedisCache{client: client, ttl: ttl}, nil
}

func (c *RedisCache) Client() *redis.Client {
	return c.client
}

// GetURL retrieves the original URL for a short code from cache.
func (c *RedisCache) GetURL(ctx context.Context, shortCode string) (string, error) {
	key := cacheKey(shortCode)
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// SetURL caches the mapping from short code to original URL.
func (c *RedisCache) SetURL(ctx context.Context, shortCode, originalURL string, expireAt *time.Time) error {
	key := cacheKey(shortCode)
	ttl := c.ttl
	if expireAt != nil {
		customTTL := time.Until(*expireAt)
		if customTTL > 0 && customTTL < ttl {
			ttl = customTTL
		}
	}
	return c.client.Set(ctx, key, originalURL, ttl).Err()
}

// DeleteURL removes a short code from cache.
func (c *RedisCache) DeleteURL(ctx context.Context, shortCode string) error {
	return c.client.Del(ctx, cacheKey(shortCode)).Err()
}

// CheckRateLimit implements a token bucket rate limiter using Redis Lua script.
// Returns true if the request is allowed, false if rate limited.
func (c *RedisCache) CheckRateLimit(ctx context.Context, key string, rate, burst int, window time.Duration) (bool, error) {
	script := redis.NewScript(`
		local key = KEYS[1]
		local rate = tonumber(ARGV[1])
		local burst = tonumber(ARGV[2])
		local window = tonumber(ARGV[3])
		local now = tonumber(ARGV[4])

		local tokens = redis.call("HGET", key, "tokens")
		local last = redis.call("HGET", key, "last")

		if tokens == false then
			tokens = burst
			last = now
		else
			local elapsed = now - tonumber(last)
			local refill = math.floor(elapsed * rate / window)
			tokens = math.min(burst, tonumber(tokens) + refill)
		end

		if tokens < 1 then
			redis.call("EXPIRE", key, math.ceil(window / rate))
			return 0
		end

		redis.call("HSET", key, "tokens", tokens - 1, "last", now)
		redis.call("EXPIRE", key, math.ceil(burst * window / rate) + 1)
		return 1
	`)

	now := time.Now().UnixMilli()
	result, err := script.Run(ctx, c.client, []string{key},
		rate, burst, window.Milliseconds(), now,
	).Int()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

// IncrementClickCount atomically increments the click counter for a link.
func (c *RedisCache) IncrementClickCount(ctx context.Context, linkID int64) error {
	key := fmt.Sprintf("clicks:%d", linkID)
	return c.client.Incr(ctx, key).Err()
}

// GetClickCount returns the click count from Redis cache.
func (c *RedisCache) GetClickCount(ctx context.Context, linkID int64) (int64, error) {
	key := fmt.Sprintf("clicks:%d", linkID)
	return c.client.Get(ctx, key).Int64()
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

func cacheKey(shortCode string) string {
	return "url:" + shortCode
}
