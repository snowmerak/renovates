package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type WebhookNotifier struct {
	URL string
}

func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{URL: url}
}

type webhookPayload struct {
	Repo   string `json:"repo"`
	Result string `json:"result"`
}

func (n *WebhookNotifier) Notify(ctx context.Context, repo string, result string) error {
	if n.URL == "" {
		return nil
	}

	payload := webhookPayload{
		Repo:   repo,
		Result: result,
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
