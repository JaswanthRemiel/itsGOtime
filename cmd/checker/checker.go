package main

import (
	"context"
	"net/http"
	"time"
)

// now returns the current time using time.Now().
// This function wraps time.Now() to allow for potential mocking in tests.
func now() time.Time {
	return time.Now()
}

// shouldCheck determines if a target should be checked based on its last check history.
// It compares the timestamp of the most recent history point against the configured interval.
// Returns true if:
//   - The history is empty (no previous checks exist)
//   - The last check timestamp cannot be parsed
//   - The time since the last check exceeds or equals the specified interval in seconds
//
// This ensures targets are only checked at their configured polling intervals.
func shouldCheck(last []HistoryPoint, intervalSec int) bool {
	if intervalSec <= 0 {
		intervalSec = 60
	}
	if len(last) == 0 {
		return true
	}
	latest := last[len(last)-1]
	t, err := time.Parse(time.RFC3339, latest.Timestamp)
	if err != nil {
		return true
	}
	next := t.Add(time.Duration(intervalSec) * time.Second)
	return now().After(next) || now().Equal(next)
}

// performCheck executes an HTTP health check against the specified target.
// It sends an HTTP request using the configured method (defaults to GET) and timeout
// (defaults to 10 seconds). If retries are configured, failed attempts will be retried
// with a 200ms delay between attempts.
//
// The function returns a CheckResult containing:
//   - Target name and URL
//   - Timestamp of the check
//   - Availability status (up/down) based on expected status code
//   - Actual HTTP status code received
//   - Response latency in milliseconds
//   - Any error message if the request failed
//
// A target is considered "up" if the response status code matches the expected status
// (or if no expected status is configured, any successful response counts as up).
func performCheck(t Target) (CheckResult, error) {
	timeout := time.Duration(10) * time.Second
	if t.TimeoutSeconds > 0 {
		timeout = time.Duration(t.TimeoutSeconds) * time.Second
	}
	method := t.Method
	if method == "" {
		method = "GET"
	}
	client := &http.Client{Timeout: timeout}

	var lastErr string
	var finalStatus int
	var latency time.Duration

	attempts := 1
	if t.Retries > 0 {
		attempts = 1 + t.Retries
	}

	for i := 0; i < attempts; i++ {
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		req, _ := http.NewRequestWithContext(ctx, method, t.URL, nil)
		resp, err := client.Do(req)
		latency = time.Since(start)
		if err != nil {
			lastErr = err.Error()
			cancel()
			if i < attempts-1 {
				time.Sleep(200 * time.Millisecond)
				continue
			}
			return CheckResult{
				Name:      t.Name,
				URL:       t.URL,
				Timestamp: now().Format(time.RFC3339),
				Up:        false,
				Status:    0,
				LatencyMs: latency.Milliseconds(),
				Error:     lastErr,
			}, nil
		}
		finalStatus = resp.StatusCode
		_ = resp.Body.Close()
		cancel()
		break
	}

	up := finalStatus != 0 && (t.ExpectStatus == 0 || finalStatus == t.ExpectStatus)

	return CheckResult{
		Name:      t.Name,
		URL:       t.URL,
		Timestamp: now().Format(time.RFC3339),
		Up:        up,
		Status:    finalStatus,
		LatencyMs: latency.Milliseconds(),
	}, nil
}
