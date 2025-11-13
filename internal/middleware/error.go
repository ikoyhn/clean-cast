package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	appErrors "ikoyhn/podcast-sponsorblock/internal/errors"

	"github.com/labstack/echo/v4"
	"ikoyhn/podcast-sponsorblock/internal/logger"
)

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey = "request_id"
)

// ErrorResponse represents the JSON structure for error responses
type ErrorResponse struct {
	Code      appErrors.ErrorCode    `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}

// ErrorHandler creates a custom error handler middleware for Echo
func ErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		// Don't process if response has already been committed
		if c.Response().Committed {
			return
		}

		// Get request ID from context
		requestID := getRequestID(c)

		var appErr *appErrors.AppError
		var echoErr *echo.HTTPError
		var httpStatus int
		var errorResponse ErrorResponse

		// Determine error type and create appropriate response
		switch {
		case err == nil:
			return

		// Check if it's our custom AppError
		case appErrors.As(err, &appErr):
			httpStatus = appErr.HTTPStatus
			errorResponse = ErrorResponse{
				Code:      appErr.Code,
				Message:   appErr.Message,
				Details:   appErr.Details,
				RequestID: requestID,
			}

			// Log the error with underlying error if present
			if appErr.Err != nil {
				c.Logger().Errorf("[%s] %s: %v", requestID, appErr.Code, appErr.Err)
			} else {
				c.Logger().Warnf("[%s] %s: %s", requestID, appErr.Code, appErr.Message)
			}

		// Check if it's an Echo HTTPError
		case appErrors.As(err, &echoErr):
			httpStatus = echoErr.Code
			message := fmt.Sprintf("%v", echoErr.Message)

			// Map Echo errors to our error codes
			code := mapHTTPStatusToErrorCode(httpStatus)

			errorResponse = ErrorResponse{
				Code:      code,
				Message:   message,
				RequestID: requestID,
			}

			// Add internal error message if present
			if echoErr.Internal != nil {
				if errorResponse.Details == nil {
					errorResponse.Details = make(map[string]interface{})
				}
				errorResponse.Details["internal"] = echoErr.Internal.Error()
				c.Logger().Errorf("[%s] Echo error: %v (internal: %v)", requestID, echoErr.Message, echoErr.Internal)
			} else {
				c.Logger().Warnf("[%s] Echo error: %v", requestID, echoErr.Message)
			}

		// Handle generic errors
		default:
			httpStatus = http.StatusInternalServerError
			errorResponse = ErrorResponse{
				Code:      appErrors.ErrCodeInternal,
				Message:   "An internal server error occurred",
				RequestID: requestID,
			}
			c.Logger().Errorf("[%s] Unexpected error: %v", requestID, err)
		}

		// Send JSON response
		if err := c.JSON(httpStatus, errorResponse); err != nil {
			c.Logger().Errorf("[%s] Failed to send error response: %v", requestID, err)
		}
	}
}

// RecoverMiddleware recovers from panics and converts them to structured errors
func RecoverMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					// Get request ID
					requestID := getRequestID(c)

					// Get stack trace
					stack := debug.Stack()

					// Log the panic with stack trace
					c.Logger().Errorf("[%s] PANIC RECOVERED: %v\n%s", requestID, r, stack)

					// Create error response
					err := appErrors.NewInternalError("A panic occurred while processing your request").
						WithRequestID(requestID).
						WithDetail("panic", fmt.Sprintf("%v", r))

					// Don't include stack trace in production, but log it
					if c.Echo().Debug {
						err.WithDetail("stack", string(stack))
					}

					// Use the error handler to send response
					c.Error(err)
				}
			}()
			return next(c)
		}
	}
}

// ValidationErrorMiddleware provides helper for creating validation errors
func ValidationErrorMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Add validation helper to context
			c.Set("validation_errors", make(map[string]string))
			return next(c)
		}
	}
}

// mapHTTPStatusToErrorCode maps HTTP status codes to our error codes
func mapHTTPStatusToErrorCode(status int) appErrors.ErrorCode {
	switch status {
	case http.StatusBadRequest:
		return appErrors.ErrCodeBadRequest
	case http.StatusUnauthorized:
		return appErrors.ErrCodeUnauthorized
	case http.StatusForbidden:
		return appErrors.ErrCodeForbidden
	case http.StatusNotFound:
		return appErrors.ErrCodeNotFound
	case http.StatusConflict:
		return appErrors.ErrCodeConflict
	case http.StatusRequestTimeout:
		return appErrors.ErrCodeTimeout
	case http.StatusTooManyRequests:
		return appErrors.ErrCodeRateLimit
	case http.StatusInternalServerError:
		return appErrors.ErrCodeInternal
	case http.StatusBadGateway:
		return appErrors.ErrCodeExternalService
	default:
		return appErrors.ErrCodeInternal
	}
}

// getRequestID retrieves the request ID from context
func getRequestID(c echo.Context) string {
	if id, ok := c.Get(RequestIDKey).(string); ok {
		return id
	}
	return "unknown"
}

// Helper functions for controllers

// ReturnError is a helper to return structured errors from handlers
func ReturnError(c echo.Context, err error) error {
	if appErr, ok := err.(*appErrors.AppError); ok {
		// Add request ID if not already set
		if appErr.RequestID == "" {
			appErr.WithRequestID(getRequestID(c))
		}
		return c.JSON(appErr.HTTPStatus, ErrorResponse{
			Code:      appErr.Code,
			Message:   appErr.Message,
			Details:   appErr.Details,
			RequestID: appErr.RequestID,
		})
	}

	// Wrap as internal error if not already an AppError
	appErr := appErrors.WrapInternalError(err, "An error occurred").
		WithRequestID(getRequestID(c))

	return c.JSON(appErr.HTTPStatus, ErrorResponse{
		Code:      appErr.Code,
		Message:   appErr.Message,
		Details:   appErr.Details,
		RequestID: appErr.RequestID,
	})
}

// LogAndReturnError logs an error and returns a structured error response
func LogAndReturnError(c echo.Context, err error, logLevel logger.Logger.Lvl) error {
	requestID := getRequestID(c)

	switch logLevel {
	case logger.Logger.DEBUG:
		c.Logger().Debugf("[%s] %v", requestID, err)
	case logger.Logger.INFO:
		c.Logger().Infof("[%s] %v", requestID, err)
	case logger.Logger.WARN:
		c.Logger().Warnf("[%s] %v", requestID, err)
	case logger.Logger.ERROR:
		c.Logger().Errorf("[%s] %v", requestID, err)
	}

	return ReturnError(c, err)
}
