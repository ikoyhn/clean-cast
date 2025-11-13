package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"
)

// RequestIDMiddleware generates a unique request ID for each request
func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if request ID is already set in header
			requestID := c.Request().Header.Get(RequestIDHeader)

			// Generate new request ID if not present
			if requestID == "" {
				requestID = generateRequestID()
			}

			// Set request ID in context
			c.Set(RequestIDKey, requestID)

			// Set request ID in response header
			c.Response().Header().Set(RequestIDHeader, requestID)

			// Continue with next handler
			return next(c)
		}
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// Generate 16 random bytes
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}

	// Convert to hex string
	return hex.EncodeToString(b)
}

// GetRequestID retrieves the request ID from the Echo context
func GetRequestID(c echo.Context) string {
	if id, ok := c.Get(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// SetRequestID sets the request ID in the Echo context
func SetRequestID(c echo.Context, requestID string) {
	c.Set(RequestIDKey, requestID)
	c.Response().Header().Set(RequestIDHeader, requestID)
}
