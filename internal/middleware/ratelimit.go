package middleware

import (
	appErrors "ikoyhn/podcast-sponsorblock/internal/errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// IPRateLimiter holds rate limiters for each IP address
type IPRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	r        rate.Limit
	b        int
}

// NewIPRateLimiter creates a new IP rate limiter
// r is the rate (requests per second)
// b is the burst size
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
	}

	// Cleanup stale entries every 5 minutes
	go i.cleanupStaleEntries()

	return i
}

// GetLimiter returns a rate limiter for the given IP
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.limiters[ip] = limiter
	}

	return limiter
}

// cleanupStaleEntries removes limiters that haven't been used recently
func (i *IPRateLimiter) cleanupStaleEntries() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		i.mu.Lock()
		// Remove all entries - they will be recreated if still needed
		// This prevents memory leaks from IPs that no longer make requests
		i.limiters = make(map[string]*rate.Limiter)
		i.mu.Unlock()
	}
}

// getIP extracts the IP address from the request
func getIP(c echo.Context) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	forwarded := c.Request().Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, use the first one
		ips := parseForwardedFor(forwarded)
		if len(ips) > 0 {
			return ips[0]
		}
	}

	// Check X-Real-IP header
	realIP := c.Request().Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request().RemoteAddr)
	if err != nil {
		return c.Request().RemoteAddr
	}
	return ip
}

// parseForwardedFor parses the X-Forwarded-For header
func parseForwardedFor(forwarded string) []string {
	var ips []string
	for _, ip := range splitAndTrim(forwarded, ",") {
		if ip != "" {
			ips = append(ips, ip)
		}
	}
	return ips
}

// splitAndTrim splits a string by separator and trims whitespace
func splitAndTrim(s, sep string) []string {
	var result []string
	for _, item := range splitString(s, sep) {
		trimmed := trimWhitespace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// splitString splits a string by separator
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

// trimWhitespace removes leading and trailing whitespace
func trimWhitespace(s string) string {
	start := 0
	end := len(s)

	// Trim leading whitespace
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	// Trim trailing whitespace
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

// RateLimitMiddleware creates rate limiting middleware
// requestsPerMinute is the number of requests allowed per minute per IP
func RateLimitMiddleware(requestsPerMinute float64) echo.MiddlewareFunc {
	// Convert requests per minute to requests per second
	requestsPerSecond := requestsPerMinute / 60.0

	// Burst allows brief bursts of traffic
	// Set burst to allow a few requests at once (e.g., 1/6 of the per-minute limit)
	burst := int(requestsPerMinute / 6)
	if burst < 1 {
		burst = 1
	}

	limiter := NewIPRateLimiter(rate.Limit(requestsPerSecond), burst)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := getIP(c)
			limiter := limiter.GetLimiter(ip)

			if !limiter.Allow() {
				return appErrors.NewRateLimitError("Rate limit exceeded").
					WithDetail("ip", ip)
			}

			return next(c)
		}
	}
}
