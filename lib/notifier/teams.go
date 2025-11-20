package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/snowmerak/renovates/lib/renovate"
)

type TeamsNotifier struct {
	URL string
}

func NewTeamsNotifier(url string) *TeamsNotifier {
	return &TeamsNotifier{URL: url}
}

func (n *TeamsNotifier) Notify(ctx context.Context, repo string, updates []renovate.UpdateInfo) error {
	if n.URL == "" {
		return nil
	}

	if len(updates) == 0 {
		return nil
	}

	// Construct the rows
	var rows []interface{}
	for _, u := range updates {
		updateTypeColor := "Default"
		updateTypeText := u.UpdateType

		// Simple mapping for visual flair
		switch u.UpdateType {
		case "major":
			updateTypeColor = "Attention"
			updateTypeText = "🚨 " + u.UpdateType
		case "minor":
			updateTypeColor = "Warning"
			updateTypeText = "⚠️ " + u.UpdateType
		case "patch":
			updateTypeColor = "Good"
			updateTypeText = "✅ " + u.UpdateType
		}

		row := map[string]interface{}{
			"type":      "ColumnSet",
			"separator": true,
			"columns": []interface{}{
				map[string]interface{}{
					"type":  "Column",
					"width": "stretch",
					"items": []interface{}{
						map[string]interface{}{"type": "TextBlock", "text": u.DepName, "wrap": true, "size": "Small"},
					},
				},
				map[string]interface{}{
					"type":  "Column",
					"width": "auto",
					"items": []interface{}{
						map[string]interface{}{"type": "TextBlock", "text": fmt.Sprintf("%s → %s", u.CurrentVersion, u.NewVersion), "size": "Small"},
					},
				},
				map[string]interface{}{
					"type":  "Column",
					"width": "60px",
					"items": []interface{}{
						map[string]interface{}{"type": "TextBlock", "text": updateTypeText, "color": updateTypeColor, "size": "Small", "horizontalAlignment": "Right", "weight": "Bolder"},
					},
				},
			},
		}
		rows = append(rows, row)
	}

	cardBody := []interface{}{
		map[string]interface{}{
			"type":   "TextBlock",
			"text":   "📢 Dependency Updates",
			"weight": "Bolder",
			"size":   "Large",
			"color":  "Accent",
		},
		map[string]interface{}{
			"type":     "TextBlock",
			"text":     fmt.Sprintf("새로운 의존성 업데이트가 감지되었습니다. (%s)", repo),
			"isSubtle": true,
			"wrap":     true,
		},
		map[string]interface{}{
			"type":  "Container",
			"style": "emphasis",
			"items": []interface{}{
				map[string]interface{}{
					"type": "ColumnSet",
					"columns": []interface{}{
						map[string]interface{}{
							"type":  "Column",
							"width": "stretch",
							"items": []interface{}{
								map[string]interface{}{"type": "TextBlock", "text": "📦 패키지명", "weight": "Bolder", "size": "Small"},
							},
						},
						map[string]interface{}{
							"type":  "Column",
							"width": "auto",
							"items": []interface{}{
								map[string]interface{}{"type": "TextBlock", "text": "버전 변경", "weight": "Bolder", "size": "Small"},
							},
						},
						map[string]interface{}{
							"type":  "Column",
							"width": "60px",
							"items": []interface{}{
								map[string]interface{}{"type": "TextBlock", "text": "유형", "weight": "Bolder", "size": "Small", "horizontalAlignment": "Right"},
							},
						},
					},
				},
			},
		},
		map[string]interface{}{
			"type":  "Container",
			"id":    "UpdateListContainer",
			"items": rows,
		},
	}

	payload := map[string]interface{}{
		"type": "message",
		"attachments": []interface{}{
			map[string]interface{}{
				"contentType": "application/vnd.microsoft.card.adaptive",
				"contentUrl":  nil,
				"content": map[string]interface{}{
					"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
					"type":    "AdaptiveCard",
					"version": "1.5",
					"body":    cardBody,
					"actions": []interface{}{
						map[string]interface{}{
							"type":  "Action.OpenUrl",
							"title": "🔗 GitHub Repo 바로가기",
							"url":   fmt.Sprintf("https://github.com/%s", repo),
						},
					},
					"msteams": map[string]interface{}{
						"width": "Full",
					},
				},
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal teams payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.URL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create teams request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send teams notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("teams webhook failed with status code: %d", resp.StatusCode)
	}

	return nil
}
