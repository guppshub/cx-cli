package connection

import (
	"strings"
	"time"
)

// RestartPolicy governs whether and when a failed connection should be retried.
type RestartPolicy interface {
	// ShouldRetry returns true if the connection should be retried given the error.
	ShouldRetry(err error) bool
	// NextDelay returns the duration to wait before the next retry attempt.
	NextDelay() time.Duration
	// Reset resets the internal retry counter and delay back to initial values.
	Reset()
}

// permanentErrorPatterns contains substrings that indicate non-retryable errors.
var permanentErrorPatterns = []string{
	"AccessDenied",
	"InvalidInstanceId",
	"TargetNotConnected",
	"ExpiredToken",
	"ExpiredTokenException",
	"InvalidIdentityToken",
	"UnrecognizedClientException",
	"AuthFailure",
	"invalid profile",
	"NoCredentialProviders",
	"could not find profile",
}

// IsPermanentError checks whether the given error is a non-retryable, permanent failure.
func IsPermanentError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, pattern := range permanentErrorPatterns {
		if strings.Contains(msg, pattern) {
			return true
		}
	}
	return false
}

// FixedBackoff implements RestartPolicy with a constant delay and a maximum retry count.
type FixedBackoff struct {
	Delay      time.Duration
	MaxRetries int
	attempt    int
}

// NewFixedBackoff creates a FixedBackoff policy.
func NewFixedBackoff(delay time.Duration, maxRetries int) *FixedBackoff {
	return &FixedBackoff{Delay: delay, MaxRetries: maxRetries}
}

// ShouldRetry returns true if the error is transient and the retry limit has not been exceeded.
func (f *FixedBackoff) ShouldRetry(err error) bool {
	if IsPermanentError(err) {
		return false
	}
	if f.MaxRetries > 0 && f.attempt >= f.MaxRetries {
		return false
	}
	f.attempt++
	return true
}

// NextDelay returns the fixed delay duration.
func (f *FixedBackoff) NextDelay() time.Duration {
	return f.Delay
}

// Reset resets the attempt counter.
func (f *FixedBackoff) Reset() {
	f.attempt = 0
}

// NoRetry implements RestartPolicy that never retries.
type NoRetry struct{}

// ShouldRetry always returns false.
func (n *NoRetry) ShouldRetry(_ error) bool {
	return false
}

// NextDelay returns zero.
func (n *NoRetry) NextDelay() time.Duration {
	return 0
}

// Reset is a no-op.
func (n *NoRetry) Reset() {}
