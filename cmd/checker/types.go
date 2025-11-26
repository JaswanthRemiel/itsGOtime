package main

// Target represents a single monitoring target configuration.
// It holds all the settings needed to perform an HTTP health check on a URL,
// including the target name, URL, HTTP method, expected status code,
// retry count, timeout duration, and polling interval.
type Target struct {
	Name            string `yaml:"name"`
	URL             string `yaml:"url"`
	Method          string `yaml:"method"`
	ExpectStatus    int    `yaml:"expect_status"`
	Retries         int    `yaml:"retries"`
	TimeoutSeconds  int    `yaml:"timeout_seconds"`
	IntervalSeconds int    `yaml:"interval_seconds"`
}

// Config represents the overall monitoring configuration loaded from monitors.yaml.
// It contains the global interval setting and a list of targets to monitor.
type Config struct {
	IntervalSeconds int      `yaml:"interval_seconds"`
	Targets         []Target `yaml:"targets"`
}

// CheckResult represents the outcome of a single health check on a target.
// It includes the target's name, URL, timestamp of the check, availability status,
// HTTP status code, response latency in milliseconds, and any error message.
type CheckResult struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	Timestamp string `json:"timestamp"`
	Up        bool   `json:"up"`
	Status    int    `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
	Error     string `json:"error,omitempty"`
}

// HistoryPoint represents a single point in the monitoring history.
// It captures the timestamp and availability status (up or down) at that moment.
type HistoryPoint struct {
	Timestamp string `json:"timestamp"`
	Up        bool   `json:"up"`
}

// History is a map that stores the monitoring history for each target.
// The key is the target name and the value is a slice of HistoryPoint entries.
type History map[string][]HistoryPoint
