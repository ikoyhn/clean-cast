package app

import (
	"time"

	"fmt"

	"github.com/labstack/echo/v4"
	log "github.com/labstack/gommon/log"
)

func setupLogging(e *echo.Echo) {
	log.SetLevel(log.INFO)
	log.SetHeader("${time_rfc3339} | ${level} | ${message}")

	e.Use(redactTokenLogger)
}

func redactTokenLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		stop := time.Now()

		req := c.Request()
		res := c.Response()

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

		// Log format
		fmt.Printf("%s | %s %s | %d | %v\n",
			stop.Format(time.RFC3339),
			req.Method,
			redactedURI,
			res.Status,
			stop.Sub(start),
		)

		return err
	}
}
