package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/snowmerak/renovates/internal/models"
)

type TeamsNotifier struct {
	WebhookURL string
}

func NewTeamsNotifier(url string) *TeamsNotifier {
	return &TeamsNotifier{WebhookURL: url}
}

func (n *TeamsNotifier) Notify(result models.RenovateResult) error {
	color := "00FF00" // Green
	if !result.Success {
		color = "FF0000" // Red
	}

	payload := map[string]interface{}{
		"@type":      "MessageCard",
		"@context":   "http://schema.org/extensions",
		"themeColor": color,
		"summary":    fmt.Sprintf("Renovate Result: %s", result.Repository),
		"sections": []map[string]interface{}{
			{
				"activityTitle":    fmt.Sprintf("Renovate Report: %s", result.Repository),
				"activitySubtitle": fmt.Sprintf("Status: %v", result.Success),
				"text":             fmt.Sprintf("Output length: %d chars", len(result.Output)),
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal teams payload: %w", err)
	}

	resp, err := http.Post(n.WebhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send teams webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("teams webhook failed with status: %d", resp.StatusCode)
	}

	return nil
}
