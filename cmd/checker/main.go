package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Target struct {
	Name            string `yaml:"name"`
	URL             string `yaml:"url"`
	Method          string `yaml:"method"`
	ExpectStatus    int    `yaml:"expect_status"`
	Retries         int    `yaml:"retries"`
	TimeoutSeconds  int    `yaml:"timeout_seconds"`
	IntervalSeconds int    `yaml:"interval_seconds"`
}

type Config struct {
	IntervalSeconds int      `yaml:"interval_seconds"`
	Targets         []Target `yaml:"targets"`
}

type CheckResult struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	Timestamp string `json:"timestamp"`
	Up        bool   `json:"up"`
	Status    int    `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
	Error     string `json:"error,omitempty"`
}

type HistoryPoint struct {
	Timestamp string `json:"timestamp"`
	Up        bool   `json:"up"`
}

type History map[string][]HistoryPoint

func loadConfig(path string) (Config, error) {
	var cfg Config
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	if cfg.IntervalSeconds == 0 {
		cfg.IntervalSeconds = 60
	}
	return cfg, nil
}

func loadHistory(path string) (History, error) {
	h := History{}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return h, nil
		}
		return nil, err
	}
	if len(b) == 0 {
		return h, nil
	}
	if err := json.Unmarshal(b, &h); err != nil {
		return nil, err
	}
	return h, nil
}

func saveJSON(path string, v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}
	return ioutil.WriteFile(path, b, 0644)
}

func now() time.Time {
	return time.Now()
}

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

func limitHistoryPoints(points []HistoryPoint, limit int) []HistoryPoint {
	if len(points) <= limit {
		return points
	}
	return points[len(points)-limit:]
}

func main() {
	cfg, err := loadConfig("monitors.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load monitors.yaml: %v\n", err)
		os.Exit(2)
	}
	historyPath := filepath.Join("gh-pages", "history.json")
	history, err := loadHistory(historyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load history: %v\n", err)
		os.Exit(2)
	}
	results := []CheckResult{}
	for _, t := range cfg.Targets {
		interval := cfg.IntervalSeconds
		if t.IntervalSeconds > 0 {
			interval = t.IntervalSeconds
		}
		lastPoints := history[t.Name]
		if shouldCheck(lastPoints, interval) {
			res, err := performCheck(t)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error checking %s: %v\n", t.Name, err)
				continue
			}
			results = append(results, res)
			history[t.Name] = append(history[t.Name], HistoryPoint{
				Timestamp: res.Timestamp,
				Up:        res.Up,
			})
			// Keep 24 hours of history
			// 24 hours * 3600 seconds / interval = number of points
			maxPoints := 86400 / interval
			if maxPoints < 1 {
				maxPoints = 1
			}
			history[t.Name] = limitHistoryPoints(history[t.Name], maxPoints)
		} else {
			if len(lastPoints) > 0 {
				latest := lastPoints[len(lastPoints)-1]
				results = append(results, CheckResult{
					Name:      t.Name,
					URL:       t.URL,
					Timestamp: now().Format(time.RFC3339),
					Up:        latest.Up,
					Status:    0,
					LatencyMs: 0,
				})
			} else {
				res, err := performCheck(t)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error checking %s: %v\n", t.Name, err)
					continue
				}
				results = append(results, res)
				history[t.Name] = append(history[t.Name], HistoryPoint{
					Timestamp: res.Timestamp,
					Up:        res.Up,
				})
				// Keep 24 hours of history
				maxPoints := 86400 / interval
				if maxPoints < 1 {
					maxPoints = 1
				}
				history[t.Name] = limitHistoryPoints(history[t.Name], maxPoints)
			}
		}
	}

	status := struct {
		GeneratedAt string        `json:"generated_at"`
		Results     []CheckResult `json:"results"`
	}{
		GeneratedAt: now().Format(time.RFC3339),
		Results:     results,
	}

	if err := saveJSON("status.json", status); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write status.json: %v\n", err)
		os.Exit(2)
	}

	if err := saveJSON(historyPath, history); err != nil {
		_ = saveJSON("history.json", history)
	}
}
