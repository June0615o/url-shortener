package middleware

import (
	"net/http"
	"time"

	"github.com/panhao/url-shortener/internal/cache"
)

// RateLimitRedis returns a middleware that uses Redis token bucket for rate limiting.
func RateLimitRedis(redisCache *cache.RedisCache, requestsPerSecond int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getIP(r)
			key := "ratelimit:ip:" + ip

			allowed, err := redisCache.CheckRateLimit(
				r.Context(), key,
				requestsPerSecond,           // rate: tokens per window
				requestsPerSecond*2,          // burst
				time.Second,                  // window
			)

			if err != nil || !allowed {
				if err != nil {
					// Redis error — allow through (fail open)
					// In production you might want to fail closed
				} else {
					http.Error(w, `{"error":"Rate limit exceeded"}`, http.StatusTooManyRequests)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitCreate returns a stricter rate limiter for the link creation endpoint.
func RateLimitCreate(redisCache *cache.RedisCache, requestsPerSecond int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getIP(r)
			key := "ratelimit:create:" + ip

			allowed, err := redisCache.CheckRateLimit(
				r.Context(), key,
				requestsPerSecond,
				requestsPerSecond,
				time.Second,
			)

			if err != nil || !allowed {
				if err == nil {
					http.Error(w, `{"error":"Too many requests, please slow down"}`, http.StatusTooManyRequests)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
