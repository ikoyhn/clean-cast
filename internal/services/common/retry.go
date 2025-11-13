package common

import (
	"context"
	"fmt"
	"math"
	"time"

	"ikoyhn/podcast-sponsorblock/internal/logger"
)

// RetryConfig holds the configuration for retry behavior
type RetryConfig struct {
	MaxAttempts     int           // Maximum number of retry attempts
	InitialDelay    time.Duration // Initial delay before first retry
	MaxDelay        time.Duration // Maximum delay between retries
	Multiplier      float64       // Multiplier for exponential backoff
	OnRetry         func(attempt int, err error) // Callback on retry
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		OnRetry:      nil,
	}
}

// DownloadRetryConfig returns a retry configuration optimized for downloads
func DownloadRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 2 * time.Second,
		MaxDelay:     60 * time.Second,
		Multiplier:   2.0,
		OnRetry: func(attempt int, err error) {
			logger.Logger.Warn().
				Int("attempt", attempt).
				Err(err).
				Msg("Download failed, retrying...")
		},
	}
}

// APIRetryConfig returns a retry configuration optimized for API calls
func APIRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 500 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		OnRetry: func(attempt int, err error) {
			logger.Logger.Warn().
				Int("attempt", attempt).
				Err(err).
				Msg("API call failed, retrying...")
		},
	}
}

// RetryWithBackoff executes a function with exponential backoff retry logic
// Returns the last error if all attempts fail
func RetryWithBackoff(ctx context.Context, config RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Execute the function
		err := fn()
		if err == nil {
			// Success
			return nil
		}

		lastErr = err

		// If this was the last attempt, return the error
		if attempt == config.MaxAttempts {
			logger.Logger.Error().
				Err(lastErr).
				Int("attempts", attempt).
				Msg("All retry attempts exhausted")
			return fmt.Errorf("all retry attempts failed after %d tries: %w", attempt, lastErr)
		}

		// Calculate delay with exponential backoff
		delay := calculateDelay(attempt-1, config.InitialDelay, config.MaxDelay, config.Multiplier)

		// Call retry callback if provided
		if config.OnRetry != nil {
			config.OnRetry(attempt, err)
		}

		logger.Logger.Debug().
			Int("attempt", attempt).
			Int("max_attempts", config.MaxAttempts).
			Dur("delay", delay).
			Msg("Retrying after delay")

		// Wait for the calculated delay or context cancellation
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		}
	}

	return lastErr
}

// RetryWithBackoffResult executes a function with exponential backoff retry logic and returns a result
func RetryWithBackoffResult[T any](ctx context.Context, config RetryConfig, fn func() (T, error)) (T, error) {
	var result T
	var lastErr error

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Execute the function
		res, err := fn()
		if err == nil {
			// Success
			return res, nil
		}

		lastErr = err

		// If this was the last attempt, return the error
		if attempt == config.MaxAttempts {
			logger.Logger.Error().
				Err(lastErr).
				Int("attempts", attempt).
				Msg("All retry attempts exhausted")
			return result, fmt.Errorf("all retry attempts failed after %d tries: %w", attempt, lastErr)
		}

		// Calculate delay with exponential backoff
		delay := calculateDelay(attempt-1, config.InitialDelay, config.MaxDelay, config.Multiplier)

		// Call retry callback if provided
		if config.OnRetry != nil {
			config.OnRetry(attempt, err)
		}

		logger.Logger.Debug().
			Int("attempt", attempt).
			Int("max_attempts", config.MaxAttempts).
			Dur("delay", delay).
			Msg("Retrying after delay")

		// Wait for the calculated delay or context cancellation
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			return result, fmt.Errorf("retry cancelled: %w", ctx.Err())
		}
	}

	return result, lastErr
}

// calculateDelay calculates the delay for exponential backoff
func calculateDelay(attempt int, initialDelay, maxDelay time.Duration, multiplier float64) time.Duration {
	delay := float64(initialDelay) * math.Pow(multiplier, float64(attempt))

	// Cap at maximum delay
	if delay > float64(maxDelay) {
		delay = float64(maxDelay)
	}

	return time.Duration(delay)
}

// IsRetryable returns true if the error is retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Add custom logic to determine if an error is retryable
	// For now, most errors are considered retryable
	// You can extend this with specific error type checks
	return true
}
