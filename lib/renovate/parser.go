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
	PackageFile    string `json:"packageFile"`
}

type branchInfo struct {
	BranchName string    `json:"branchName"`
	Upgrades   []upgrade `json:"upgrades"`
}

type packageFileUpdate struct {
	NewVersion string `json:"newVersion"`
	UpdateType string `json:"updateType"`
}

type packageFileDep struct {
	DepName        string              `json:"depName"`
	CurrentVersion string              `json:"currentVersion"`
	Updates        []packageFileUpdate `json:"updates"`
}

type packageFile struct {
	PackageFile string           `json:"packageFile"`
	Deps        []packageFileDep `json:"deps"`
}

type logEntry struct {
	Msg                 string                   `json:"msg"`
	BranchesInformation []branchInfo             `json:"branchesInformation"`
	Config              map[string][]packageFile `json:"config"`
}

func ParseUpdates(output []byte) string {
	lines := strings.Split(string(output), "\n")
	// Key: DepName + PackageFile + NewVersion
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
					key := fmt.Sprintf("%s|%s|%s", upgrade.DepName, upgrade.PackageFile, upgrade.NewVersion)
					msg := fmt.Sprintf("%s: %s -> %s", upgrade.DepName, upgrade.CurrentVersion, upgrade.NewVersion)
					if upgrade.PackageFile != "" {
						msg += fmt.Sprintf(" (%s)", upgrade.PackageFile)
					}
					if upgrade.UpdateType != "" {
						msg += fmt.Sprintf(" [%s]", upgrade.UpdateType)
					}
					updates[key] = msg
				}
			}
		} else if entry.Msg == "packageFiles with updates" && len(entry.Config) > 0 {
			for _, packageFiles := range entry.Config {
				for _, pf := range packageFiles {
					for _, dep := range pf.Deps {
						for _, update := range dep.Updates {
							key := fmt.Sprintf("%s|%s|%s", dep.DepName, pf.PackageFile, update.NewVersion)
							msg := fmt.Sprintf("%s: %s -> %s", dep.DepName, dep.CurrentVersion, update.NewVersion)
							if pf.PackageFile != "" {
								msg += fmt.Sprintf(" (%s)", pf.PackageFile)
							}
							if update.UpdateType != "" {
								msg += fmt.Sprintf(" [%s]", update.UpdateType)
							}
							updates[key] = msg
						}
					}
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
		sb.WriteString(fmt.Sprintf("- %s\n", updates[k]))
	}

	return sb.String()
}
