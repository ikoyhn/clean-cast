package app

import (
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/logger"
	"ikoyhn/podcast-sponsorblock/internal/metrics"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func setupLogging(e *echo.Echo) {
	// Initialize structured logger
	if os.Getenv("LOG_FORMAT") == "json" {
		logger.InitializeJSON()
	} else {
		logger.Initialize()
	}

	logger.Logger.Info().Msg("Logger initialized")

	// Add request ID middleware
	e.Use(requestIDMiddleware)

	// Add logging and metrics middleware
	e.Use(loggingAndMetricsMiddleware)
}

// requestIDMiddleware adds a unique request ID to each request
func requestIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := c.Request().Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Response().Header().Set("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		return next(c)
	}
}

// loggingAndMetricsMiddleware logs requests and records metrics
func loggingAndMetricsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		duration := time.Since(start)

		req := c.Request()
		res := c.Response()

		// Redact sensitive query parameters
		uri := *req.URL
		q := uri.Query()
		if q.Has("token") {
			q.Set("token", "[REDACTED]")
			uri.RawQuery = q.Encode()
		}
		redactedURI := uri.Path
		if uri.RawQuery != "" {
			redactedURI += "?" + uri.RawQuery
		}

		// Get request ID
		requestID, _ := c.Get("request_id").(string)

		// Record metrics
		status := strconv.Itoa(res.Status)
		endpoint := c.Path()
		if endpoint == "" {
			endpoint = redactedURI
		}
		metrics.RecordHTTPRequest(req.Method, endpoint, status, duration)

		// Structured logging
		logEvent := logger.Logger.Info()
		if res.Status >= 500 {
			logEvent = logger.Logger.Error()
		} else if res.Status >= 400 {
			logEvent = logger.Logger.Warn()
		}

		logEvent.
			Str("request_id", requestID).
			Str("method", req.Method).
			Str("uri", redactedURI).
			Int("status", res.Status).
			Dur("duration", duration).
			Str("remote_ip", c.RealIP()).
			Int64("bytes_out", res.Size).
			Msg("HTTP request")

		return err
	}
}
