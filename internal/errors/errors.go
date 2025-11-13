package errors

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// Validation errors
	ErrCodeValidation     ErrorCode = "VALIDATION_ERROR"
	ErrCodeInvalidInput   ErrorCode = "INVALID_INPUT"
	ErrCodeInvalidParam   ErrorCode = "INVALID_PARAM"
	ErrCodeInvalidFormat  ErrorCode = "INVALID_FORMAT"
	ErrCodeMissingField   ErrorCode = "MISSING_FIELD"

	// Not found errors
	ErrCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrCodeResourceNotFound ErrorCode = "RESOURCE_NOT_FOUND"
	ErrCodeFileNotFound     ErrorCode = "FILE_NOT_FOUND"

	// Server errors
	ErrCodeInternal        ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceError    ErrorCode = "SERVICE_ERROR"
	ErrCodeDatabaseError   ErrorCode = "DATABASE_ERROR"
	ErrCodeDownloadError   ErrorCode = "DOWNLOAD_ERROR"

	// Rate limiting and throttling
	ErrCodeRateLimit  ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrCodeTooMany    ErrorCode = "TOO_MANY_REQUESTS"

	// External service errors
	ErrCodeExternalService ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrCodeYouTubeError    ErrorCode = "YOUTUBE_ERROR"
	ErrCodeAPIError        ErrorCode = "API_ERROR"

	// Request errors
	ErrCodeBadRequest      ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized    ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden       ErrorCode = "FORBIDDEN"
	ErrCodeTimeout         ErrorCode = "TIMEOUT"
	ErrCodeConflict        ErrorCode = "CONFLICT"
)

// AppError represents a structured application error
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
	RequestID  string                 `json:"request_id,omitempty"`
	Err        error                  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Err.Error())
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap implements the error unwrapping interface
func (e *AppError) Unwrap() error {
	return e.Err
}

// MarshalJSON implements custom JSON marshalling
func (e *AppError) MarshalJSON() ([]byte, error) {
	type Alias AppError
	return json.Marshal(&struct {
		*Alias
		Error string `json:"error,omitempty"`
	}{
		Alias: (*Alias)(e),
		Error: func() string {
			if e.Err != nil {
				return e.Err.Error()
			}
			return ""
		}(),
	})
}

// WithRequestID adds a request ID to the error
func (e *AppError) WithRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

// WithDetail adds a single detail key-value pair to the error
func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// New creates a new AppError
func New(code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Wrap wraps an existing error with AppError
func Wrap(err error, code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

// Standard error constructors

// NewValidationError creates a validation error
func NewValidationError(message string) *AppError {
	return New(ErrCodeValidation, message, http.StatusBadRequest)
}

// NewInvalidInputError creates an invalid input error
func NewInvalidInputError(field string) *AppError {
	return New(ErrCodeInvalidInput, "Invalid input provided", http.StatusBadRequest).
		WithDetail("field", field)
}

// NewInvalidParamError creates an invalid parameter error
func NewInvalidParamError(param string) *AppError {
	return New(ErrCodeInvalidParam, "Invalid parameter provided", http.StatusBadRequest).
		WithDetail("parameter", param)
}

// NewMissingFieldError creates a missing field error
func NewMissingFieldError(field string) *AppError {
	return New(ErrCodeMissingField, "Required field is missing", http.StatusBadRequest).
		WithDetail("field", field)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return New(ErrCodeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound).
		WithDetail("resource", resource)
}

// NewResourceNotFoundError creates a resource not found error
func NewResourceNotFoundError(resourceType, resourceID string) *AppError {
	return New(ErrCodeResourceNotFound, "Resource not found", http.StatusNotFound).
		WithDetail("resource_type", resourceType).
		WithDetail("resource_id", resourceID)
}

// NewFileNotFoundError creates a file not found error
func NewFileNotFoundError(filename string) *AppError {
	return New(ErrCodeFileNotFound, "File not found", http.StatusNotFound).
		WithDetail("filename", filename)
}

// NewInternalError creates an internal server error
func NewInternalError(message string) *AppError {
	if message == "" {
		message = "An internal server error occurred"
	}
	return New(ErrCodeInternal, message, http.StatusInternalServerError)
}

// WrapInternalError wraps an error as an internal server error
func WrapInternalError(err error, message string) *AppError {
	if message == "" {
		message = "An internal server error occurred"
	}
	return Wrap(err, ErrCodeInternal, message, http.StatusInternalServerError)
}

// NewServiceError creates a service error
func NewServiceError(service string, err error) *AppError {
	return Wrap(err, ErrCodeServiceError, "Service error occurred", http.StatusInternalServerError).
		WithDetail("service", service)
}

// NewDatabaseError creates a database error
func NewDatabaseError(operation string, err error) *AppError {
	return Wrap(err, ErrCodeDatabaseError, "Database operation failed", http.StatusInternalServerError).
		WithDetail("operation", operation)
}

// NewDownloadError creates a download error
func NewDownloadError(resourceID string, err error) *AppError {
	return Wrap(err, ErrCodeDownloadError, "Failed to download resource", http.StatusInternalServerError).
		WithDetail("resource_id", resourceID)
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(message string) *AppError {
	if message == "" {
		message = "Rate limit exceeded"
	}
	return New(ErrCodeRateLimit, message, http.StatusTooManyRequests)
}

// NewExternalServiceError creates an external service error
func NewExternalServiceError(service string, err error) *AppError {
	return Wrap(err, ErrCodeExternalService, fmt.Sprintf("External service '%s' error", service), http.StatusBadGateway).
		WithDetail("service", service)
}

// NewYouTubeError creates a YouTube API error
func NewYouTubeError(err error) *AppError {
	return Wrap(err, ErrCodeYouTubeError, "YouTube service error", http.StatusBadGateway).
		WithDetail("service", "youtube")
}

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string) *AppError {
	return New(ErrCodeBadRequest, message, http.StatusBadRequest)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation string) *AppError {
	return New(ErrCodeTimeout, "Operation timed out", http.StatusRequestTimeout).
		WithDetail("operation", operation)
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *AppError {
	return New(ErrCodeConflict, message, http.StatusConflict)
}

// As is a convenience wrapper around errors.As
func As(err error, target interface{}) bool {
	return stderrors.As(err, target)
}

// Is is a convenience wrapper around errors.Is
func Is(err, target error) bool {
	return stderrors.Is(err, target)
}
