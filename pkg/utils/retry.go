package utils

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"syscall"
	"time"
)

// RetryConfig configures retry behavior for operations.
type RetryConfig struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
}

// DefaultRetryConfig returns a default retry configuration for network operations.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:  3,
		InitialWait: 1 * time.Second,
		MaxWait:     30 * time.Second,
	}
}

// IsRetryableError checks if an error is potentially transient and worth retrying.
// It detects DNS errors, connection timeouts, temporary network errors, etc.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// DNS errors
	if strings.Contains(errStr, "i/o timeout") &&
		(strings.Contains(errStr, "lookup") || strings.Contains(errStr, "dial tcp")) {
		return true
	}

	// Connection timeouts
	if strings.Contains(errStr, "connection timed out") ||
		strings.Contains(errStr, "Connection refused") ||
		strings.Contains(errStr, "no such host") {
		return true
	}

	// URL errors that are temporary
	if urlErr, ok := err.(*url.Error); ok {
		if urlErr.Temporary() {
			return true
		}
		// Check the underlying error
		if urlErr.Err != nil {
			return IsRetryableError(urlErr.Err)
		}
	}

	// Net errors that are temporary or timeouts
	if netErr, ok := err.(net.Error); ok {
		if netErr.Temporary() || netErr.Timeout() {
			return true
		}
	}

	// Syscall errors - connection reset, broken pipe, etc.
	if errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.EPIPE) ||
		errors.Is(err, syscall.ETIMEDOUT) {
		return true
	}

	return false
}

// GetRetryableErrorMessage returns a user-friendly message for retryable errors,
// helping users understand the underlying issue.
func GetRetryableErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	// DNS timeout detection
	if strings.Contains(errStr, "i/o timeout") {
		if strings.Contains(errStr, "lookup") {
			// Extract the hostname
			parts := strings.Split(errStr, "lookup ")
			if len(parts) > 1 {
				host := strings.Split(parts[1], ":")[0]
				return fmt.Sprintf("DNS resolution timeout for %s. Check your network/DNS settings.", host)
			}
			return "DNS resolution timeout. Check your network/DNS settings."
		}
		if strings.Contains(errStr, "dial tcp") {
			return "Connection timeout. The server may be unreachable or your network is down."
		}
	}

	return err.Error()
}

// RetryWithBackoff executes the given function with exponential backoff retry logic.
// It retries on transient network errors up to MaxRetries times.
func RetryWithBackoff(ctx context.Context, config RetryConfig, operation func() error) error {
	var lastErr error
	wait := config.InitialWait

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(wait):
				// Continue to next attempt
			}
			// Exponential backoff with cap
			wait *= 2
			if wait > config.MaxWait {
				wait = config.MaxWait
			}
		}

		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Only retry on specific error types
		if !IsRetryableError(err) {
			return err
		}

		// Don't log on last attempt - we'll return the error
		if attempt < config.MaxRetries {
			// Log handled by caller if needed
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", config.MaxRetries+1, lastErr)
}
