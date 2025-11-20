package renovate

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type UpdateInfo struct {
	DepName        string `json:"depName"`
	CurrentVersion string `json:"currentVersion"`
	NewVersion     string `json:"newVersion"`
	UpdateType     string `json:"updateType"`
	PackageFile    string `json:"packageFile"`
}

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

func ParseUpdates(output []byte) []UpdateInfo {
	lines := strings.Split(string(output), "\n")
	// Key: DepName + PackageFile + NewVersion
	updatesMap := make(map[string]UpdateInfo)

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
					updatesMap[key] = UpdateInfo{
						DepName:        upgrade.DepName,
						CurrentVersion: upgrade.CurrentVersion,
						NewVersion:     upgrade.NewVersion,
						UpdateType:     upgrade.UpdateType,
						PackageFile:    upgrade.PackageFile,
					}
				}
			}
		} else if entry.Msg == "packageFiles with updates" && len(entry.Config) > 0 {
			for _, packageFiles := range entry.Config {
				for _, pf := range packageFiles {
					for _, dep := range pf.Deps {
						for _, update := range dep.Updates {
							key := fmt.Sprintf("%s|%s|%s", dep.DepName, pf.PackageFile, update.NewVersion)
							updatesMap[key] = UpdateInfo{
								DepName:        dep.DepName,
								CurrentVersion: dep.CurrentVersion,
								NewVersion:     update.NewVersion,
								UpdateType:     update.UpdateType,
								PackageFile:    pf.PackageFile,
							}
						}
					}
				}
			}
		}
	}

	var updates []UpdateInfo
	for _, u := range updatesMap {
		updates = append(updates, u)
	}

	sort.Slice(updates, func(i, j int) bool {
		if updates[i].DepName != updates[j].DepName {
			return updates[i].DepName < updates[j].DepName
		}
		return updates[i].NewVersion < updates[j].NewVersion
	})

	return updates
}
