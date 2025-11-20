package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/snowmerak/renovates/internal/models"
)

type WebhookNotifier struct {
	URL string
}

func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{URL: url}
}

func (n *WebhookNotifier) Notify(result models.RenovateResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	resp, err := http.Post(n.URL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook failed with status: %d", resp.StatusCode)
	}

	return nil
}
