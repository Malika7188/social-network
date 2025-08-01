package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/Athooh/social-network/pkg/logger"
)

// RateLimiter implements a simple rate limiting middleware
type RateLimiter struct {
	requests    map[string][]time.Time
	mu          sync.Mutex
	limit       int           // Maximum number of requests
	window      time.Duration // Time window for rate limiting
	ipExtractor func(r *http.Request) string
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests:    make(map[string][]time.Time),
		limit:       limit,
		window:      window,
		ipExtractor: defaultIPExtractor,
	}
}

// defaultIPExtractor extracts the client IP from a request
func defaultIPExtractor(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}

// Middleware returns a rate limiting middleware function
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := rl.ipExtractor(r)

		rl.mu.Lock()

		// Clean up old requests
		now := time.Now()
		cutoff := now.Add(-rl.window)

		if times, exists := rl.requests[clientIP]; exists {
			var validTimes []time.Time
			for _, t := range times {
				if t.After(cutoff) {
					validTimes = append(validTimes, t)
				}
			}
			rl.requests[clientIP] = validTimes
		}

		// Check if rate limit is exceeded
		if len(rl.requests[clientIP]) >= rl.limit {
			rl.mu.Unlock()
			logger.Warn("Rate limit exceeded for IP: %s", clientIP)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"Rate limit exceeded. Please try again later."}`))
			return
		}

		// Add current request
		rl.requests[clientIP] = append(rl.requests[clientIP], now)
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
