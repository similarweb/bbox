package utils

import "time"

// Calculate exponential backoff delay
func CalculateBackoff(baseDelay time.Duration, factor int, maxDelay time.Duration, attempt int) time.Duration {
	delay := baseDelay * time.Duration(factor^attempt)
	if delay > maxDelay {
		return maxDelay
	}
	return delay
}
