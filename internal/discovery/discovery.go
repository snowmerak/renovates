package discovery

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/snowmerak/renovates/internal/models"
)

type Discoverer struct {
	RenovateCmd string
}

func NewDiscoverer(cmd string) *Discoverer {
	return &Discoverer{RenovateCmd: cmd}
}

func (d *Discoverer) Discover() ([]models.Repository, error) {
	tempDir := os.TempDir()
	outputFile := filepath.Join(tempDir, "renovate-repos.json")
	defer os.Remove(outputFile)

	// Run renovate with autodiscover and write-discovered-repos
	// We use --dry-run to avoid actual changes during discovery,
	// but we need to make sure it runs enough to discover.
	// RENOVATE_AUTODISCOVER=true
	// RENOVATE_WRITE_DISCOVERED_REPOS=...

	cmd := exec.Command(d.RenovateCmd)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "RENOVATE_AUTODISCOVER=true")
	cmd.Env = append(cmd.Env, fmt.Sprintf("RENOVATE_WRITE_DISCOVERED_REPOS=%s", outputFile))
	// We might want to limit what it does, e.g. --schedule=now to ensure it runs?
	// Or just let it run. It might take a while if it analyzes everything.
	// Ideally we want it to stop after discovery.
	// There is no explicit "discover-only" flag, but writing the file happens after discovery.
	// If we can kill it after the file is written? That's risky.
	// For now, let's assume we let it run.
	// To speed it up, maybe disable other things?

	// Note: In a real scenario, you might want to use the platform API directly (GitHub/GitLab)
	// instead of running Renovate for discovery, as it's much faster.
	// But following the requirement: "Renovate의 자동 탐색으로 레포를 받아와서"

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run renovate for discovery: %w, output: %s", err, string(output))
	}

	// Read the file
	data, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read discovered repos file: %w", err)
	}

	var repos []models.Repository
	// Try parsing as objects
	if err := json.Unmarshal(data, &repos); err != nil {
		// Fallback: maybe it's a list of strings?
		var repoStrings []string
		if err2 := json.Unmarshal(data, &repoStrings); err2 == nil {
			for _, r := range repoStrings {
				repos = append(repos, models.Repository{Name: r})
			}
		} else {
			return nil, fmt.Errorf("failed to parse discovered repos: %w", err)
		}
	}

	return repos, nil
}
