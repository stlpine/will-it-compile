package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// RateLimiter implements a simple token bucket rate limiter.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    int           // requests per window
	window  time.Duration // time window
}

type bucket struct {
	tokens     int
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	limiter := &RateLimiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
		window:  window,
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// Allow checks if a request from the given IP is allowed.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, exists := rl.buckets[ip]
	if !exists {
		b = &bucket{
			tokens:     rl.rate,
			lastRefill: time.Now(),
		}
		rl.buckets[ip] = b
	}

	// Refill tokens if window has passed
	now := time.Now()
	if now.Sub(b.lastRefill) >= rl.window {
		b.tokens = rl.rate
		b.lastRefill = now
	}

	// Check if tokens available
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// cleanup periodically removes old buckets.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, b := range rl.buckets {
			if now.Sub(b.lastRefill) > 10*time.Minute {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware returns an Echo middleware that enforces rate limiting.
func RateLimitMiddleware(limiter *RateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract IP (Echo handles X-Forwarded-For via RealIP())
			ip := c.RealIP()

			if !limiter.Allow(ip) {
				return echo.NewHTTPError(http.StatusTooManyRequests, "Rate limit exceeded")
			}

			return next(c)
		}
	}
}
