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
	Text        string            `json:"text,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackAttachment represents an attachment block in Slack with colored borders.
type SlackAttachment struct {
	Color  string       `json:"color"`
	Blocks []SlackBlock `json:"blocks"`
}

// SlackBlock represents a block inside the attachment.
type SlackBlock struct {
	Type     string            `json:"type"`
	Text     *SlackText        `json:"text,omitempty"`
	Fields   []SlackText       `json:"fields,omitempty"`
	Elements []SlackText       `json:"elements,omitempty"`
}

// SlackText represents text elements inside Slack blocks.
type SlackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// sendSlackAlert sends a structured alert to a Slack channel via an incoming webhook URL.
func sendSlackAlert(webhookURL string, res CheckResult, isRecovery bool) error {
	var color string
	var headerText string
	var statusText string

	if isRecovery {
		color = "#2eb67d" // Slack Green
		headerText = fmt.Sprintf("✅ *RECOVERY: %s is back UP*", res.Name)
		statusText = fmt.Sprintf("HTTP Status: %d", res.Status)
	} else {
		color = "#e01e5a" // Slack Red
		headerText = fmt.Sprintf("🚨 *DOWNTIME: %s is DOWN*", res.Name)
		if res.Status > 0 {
			statusText = fmt.Sprintf("HTTP Status: %d", res.Status)
		} else if res.Error != "" {
			statusText = fmt.Sprintf("Error: %s", res.Error)
		} else {
			statusText = "Unknown connection or timeout error"
		}
	}

	payload := SlackPayload{
		Text: headerText,
		Attachments: []SlackAttachment{
			{
				Color: color,
				Blocks: []SlackBlock{
					{
						Type: "section",
						Text: &SlackText{
							Type: "mrkdwn",
							Text: headerText,
						},
					},
					{
						Type: "section",
						Fields: []SlackText{
							{
								Type: "mrkdwn",
								Text: fmt.Sprintf("*Name:*\n%s", res.Name),
							},
							{
								Type: "mrkdwn",
								Text: fmt.Sprintf("*URL:*\n<%s|Link>", res.URL),
							},
							{
								Type: "mrkdwn",
								Text: fmt.Sprintf("*Status:*\n%s", statusText),
							},
							{
								Type: "mrkdwn",
								Text: fmt.Sprintf("*Response Time:*\n%d ms", res.LatencyMs),
							},
						},
					},
					{
						Type: "context",
						Elements: []SlackText{
							{
								Type: "mrkdwn",
								Text: fmt.Sprintf("Checked at: %s", res.Timestamp),
							},
						},
					},
				},
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
