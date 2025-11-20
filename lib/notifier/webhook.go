package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/snowmerak/renovates/lib/renovate"
)

type WebhookNotifier struct {
	URL string
}

func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{URL: url}
}

type webhookPayload struct {
	Repo    string                `json:"repo"`
	Updates []renovate.UpdateInfo `json:"updates"`
}

func (n *WebhookNotifier) Notify(ctx context.Context, repo string, updates []renovate.UpdateInfo) error {
	if n.URL == "" {
		return nil
	}

	payload := webhookPayload{
		Repo:    repo,
		Updates: updates,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.URL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook failed with status code: %d", resp.StatusCode)
	}

	return nil
}
