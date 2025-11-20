package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/snowmerak/renovates/lib/renovate"
)

type TelegramNotifier struct {
	Token  string
	ChatID string
}

func NewTelegramNotifier(token, chatID string) *TelegramNotifier {
	return &TelegramNotifier{
		Token:  token,
		ChatID: chatID,
	}
}

func (n *TelegramNotifier) Notify(ctx context.Context, repo string, updates []renovate.UpdateInfo) error {
	if n.Token == "" || n.ChatID == "" {
		return nil
	}

	if len(updates) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📢 *Renovate Updates for %s*\n\n", repo))

	for _, u := range updates {
		sb.WriteString(fmt.Sprintf("📦 *%s*\n", u.DepName))
		sb.WriteString(fmt.Sprintf("   %s → %s", u.CurrentVersion, u.NewVersion))
		if u.UpdateType != "" {
			sb.WriteString(fmt.Sprintf(" \\[%s\\]", u.UpdateType))
		}
		sb.WriteString("\n")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.Token)
	payload := map[string]interface{}{
		"chat_id":    n.ChatID,
		"text":       sb.String(),
		"parse_mode": "Markdown",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send telegram notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("telegram api failed with status code: %d", resp.StatusCode)
	}

	return nil
}
