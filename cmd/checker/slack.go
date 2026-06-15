package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SlackPayload represents the JSON payload structure for Slack incoming webhooks.
type SlackPayload struct {
	Attachments []SlackAttachment `json:"attachments"`
}

// SlackAttachment represents an attachment block in Slack with colored borders.
type SlackAttachment struct {
	Color     string       `json:"color,omitempty"`
	Pretext   string       `json:"pretext,omitempty"`
	Title     string       `json:"title,omitempty"`
	TitleLink string       `json:"title_link,omitempty"`
	Fields    []SlackField `json:"fields,omitempty"`
	Ts        int64        `json:"ts,omitempty"`
}

// SlackField represents a field inside the legacy attachment.
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// sendSlackAlert sends a structured alert to a Slack channel via an incoming webhook URL.
func sendSlackAlert(webhookURL string, res CheckResult, isRecovery bool) error {
	var color string
	var pretext string
	var statusText string

	if isRecovery {
		color = "#2eb67d" // Slack Green
		pretext = fmt.Sprintf("✅ *RECOVERY: %s is back UP*", res.Name)
		statusText = fmt.Sprintf("HTTP Status: %d", res.Status)
	} else {
		color = "#e01e5a" // Slack Red
		pretext = fmt.Sprintf("🚨 *DOWNTIME: %s is DOWN*", res.Name)
		if res.Status > 0 {
			statusText = fmt.Sprintf("HTTP Status: %d", res.Status)
		} else if res.Error != "" {
			statusText = fmt.Sprintf("Error: %s", res.Error)
		} else {
			statusText = "Unknown connection or timeout error"
		}
	}

	payload := SlackPayload{
		Attachments: []SlackAttachment{
			{
				Color:     color,
				Pretext:   pretext,
				Title:     res.Name,
				TitleLink: res.URL,
				Fields: []SlackField{
					{
						Title: "Status",
						Value: statusText,
						Short: true,
					},
					{
						Title: "Response Time",
						Value: fmt.Sprintf("%d ms", res.LatencyMs),
						Short: true,
					},
				},
				Ts: time.Now().Unix(),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	fmt.Printf("Alerting: Sending HTTP POST payload to Slack webhook (%d bytes)...\n", len(body))
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request to Slack: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Alerting: Slack response status: %s\n", resp.Status)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack responded with non-2xx status: %d", resp.StatusCode)
	}

	return nil
}
