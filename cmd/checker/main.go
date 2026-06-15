package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// main is the entry point of the uptime checker application.
// It performs the following operations:
//  1. Loads the monitoring configuration from monitors.yaml
//  2. Loads the existing check history from gh-pages/history.json
//  3. Iterates through each configured target and performs health checks
//     based on their individual or global interval settings
//  4. Updates the history with new check results
//  5. Writes the current status to status.json
//  6. Saves the updated history back to gh-pages/history.json
//
// The checker respects per-target intervals and only performs a check
// when the configured time has elapsed since the last check.
// History is automatically trimmed to maintain approximately 24 hours of data.
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
			handleAlerting(res, lastPoints)
			results = append(results, res)
			history[t.Name] = append(history[t.Name], HistoryPoint{
				Timestamp: res.Timestamp,
				Up:        res.Up,
			})

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
				handleAlerting(res, lastPoints)
				results = append(results, res)
				history[t.Name] = append(history[t.Name], HistoryPoint{
					Timestamp: res.Timestamp,
					Up:        res.Up,
				})
				// Keeping 24 hours of history
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

// handleAlerting checks if the status of a target transitioned and triggers Slack alerts if SLACK_WEBHOOK_URL is set.
func handleAlerting(res CheckResult, lastPoints []HistoryPoint) {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		fmt.Printf("Alerting: SLACK_WEBHOOK_URL is not set, skipping alerts for %s\n", res.Name)
		return
	}

	var prevUp bool = true
	hasHistory := len(lastPoints) > 0
	if hasHistory {
		prevUp = lastPoints[len(lastPoints)-1].Up
	}

	fmt.Printf("Alerting: Checking target %s (Up: %v, PrevUp: %v, HasHistory: %v)\n", res.Name, res.Up, prevUp, hasHistory)

	if !res.Up && (prevUp || !hasHistory) {
		fmt.Printf("Alerting: Sending downtime alert for %s...\n", res.Name)
		if err := sendSlackAlert(webhookURL, res, false); err != nil {
			fmt.Fprintf(os.Stderr, "failed to send slack downtime alert for %s: %v\n", res.Name, err)
		} else {
			fmt.Printf("Alerting: Slack downtime alert sent successfully for %s\n", res.Name)
		}
	} else if res.Up && !prevUp && hasHistory {
		fmt.Printf("Alerting: Sending recovery alert for %s...\n", res.Name)
		if err := sendSlackAlert(webhookURL, res, true); err != nil {
			fmt.Fprintf(os.Stderr, "failed to send slack recovery alert for %s: %v\n", res.Name, err)
		} else {
			fmt.Printf("Alerting: Slack recovery alert sent successfully for %s\n", res.Name)
		}
	}
}
