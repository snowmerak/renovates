package renovate

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type upgrade struct {
	DepName        string `json:"depName"`
	CurrentVersion string `json:"currentVersion"`
	NewVersion     string `json:"newVersion"`
	UpdateType     string `json:"updateType"`
}

type branchInfo struct {
	BranchName string    `json:"branchName"`
	Upgrades   []upgrade `json:"upgrades"`
}

type logEntry struct {
	Msg                 string       `json:"msg"`
	BranchesInformation []branchInfo `json:"branchesInformation"`
}

func ParseUpdates(output []byte) string {
	lines := strings.Split(string(output), "\n")
	updates := make(map[string]string)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var entry logEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if entry.Msg == "branches info extended" && len(entry.BranchesInformation) > 0 {
			for _, branch := range entry.BranchesInformation {
				for _, upgrade := range branch.Upgrades {
					msg := fmt.Sprintf("%s -> %s", upgrade.CurrentVersion, upgrade.NewVersion)
					if upgrade.UpdateType != "" {
						msg += fmt.Sprintf(" (%s)", upgrade.UpdateType)
					}
					updates[upgrade.DepName] = msg
				}
			}
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
		sb.WriteString(fmt.Sprintf("- %s: %s\n", k, updates[k]))
	}

	return sb.String()
}
