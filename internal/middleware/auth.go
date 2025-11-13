package middleware

import (
	"ikoyhn/podcast-sponsorblock/internal/config"
	appErrors "ikoyhn/podcast-sponsorblock/internal/errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// AuthMiddleware checks for TOKEN authentication if configured
func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// If no token is configured, skip authentication
			if config.Config.Token == "" {
				return next(c)
			}

			// Check for token in query parameter
			tokenParam := c.QueryParam("token")
			if tokenParam != "" {
				if tokenParam == config.Config.Token {
					return next(c)
				}
				return appErrors.New(appErrors.ErrCodeUnauthorized, "Invalid token", http.StatusUnauthorized)
			}

			// Check for token in Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" {
				// Support both "Bearer <token>" and direct token
				token := strings.TrimSpace(authHeader)
				if strings.HasPrefix(token, "Bearer ") {
					token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer"))
				}

				if token == config.Config.Token {
					return next(c)
				}
				return appErrors.New(appErrors.ErrCodeUnauthorized, "Invalid token", http.StatusUnauthorized)
			}

			// No token provided
			return appErrors.New(appErrors.ErrCodeUnauthorized, "Authentication required", http.StatusUnauthorized)
		}
	}
}
