package renovate

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type logEntry struct {
	Msg            string `json:"msg"`
	DepName        string `json:"depName"`
	NewVersion     string `json:"newVersion"`
	CurrentVersion string `json:"currentVersion"`
	UpdateType     string `json:"updateType"`
}

func ParseUpdates(output []byte) string {
	lines := strings.Split(string(output), "\n")
	updates := make(map[string]logEntry)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var entry logEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if entry.DepName != "" && entry.NewVersion != "" {
			updates[entry.DepName] = entry
		}
	}

	if len(updates) == 0 {
		return "No updates needed."
	}

	var keys []string
	for k := range updates {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString("Dependency Updates:\n")
	for _, k := range keys {
		u := updates[k]
		sb.WriteString(fmt.Sprintf("- %s: %s -> %s", u.DepName, u.CurrentVersion, u.NewVersion))
		if u.UpdateType != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", u.UpdateType))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
