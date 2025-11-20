package renovate

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/pelletier/go-toml/v2"
)

type NotifierConfig struct {
	Type   string `toml:"type"`
	URL    string `toml:"url"`
	Token  string `toml:"token"`
	ChatID string `toml:"chat_id"`
}

type DiscoveryConfig struct {
	Enabled  bool     `toml:"enabled"`
	Owner    string   `toml:"owner"`
	Topics   []string `toml:"topics"`
	Includes []string `toml:"includes"`
	Excludes []string `toml:"excludes"`
}

type Config struct {
	Command       string            `toml:"command"`
	Platform      string            `toml:"platform"`
	Token         string            `toml:"token"`
	Endpoint      string            `toml:"endpoint"`
	LogLevel      string            `toml:"log_level"`
	DryRun        bool              `toml:"dry_run"`
	Onboarding    bool              `toml:"onboarding"`
	RequireConfig string            `toml:"require_config"`
	Notifiers     []NotifierConfig  `toml:"notifiers"`
	Discovery     DiscoveryConfig   `toml:"discovery"`
	ExtraEnv      map[string]string `toml:"extra_env"`
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := toml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &cfg, nil
}

func (c *Config) ToEnv() []string {
	envs := os.Environ()

	if c.Platform != "" {
		envs = append(envs, fmt.Sprintf("RENOVATE_PLATFORM=%s", c.Platform))
	}
	if c.Token != "" {
		envs = append(envs, fmt.Sprintf("RENOVATE_TOKEN=%s", c.Token))
	}
	if c.Endpoint != "" {
		envs = append(envs, fmt.Sprintf("RENOVATE_ENDPOINT=%s", c.Endpoint))
	}

	// Enforced options
	envs = append(envs, "LOG_LEVEL=debug")
	envs = append(envs, "RENOVATE_DRY_RUN=true")
	envs = append(envs, "RENOVATE_ONBOARDING=false")
	envs = append(envs, "RENOVATE_REQUIRE_CONFIG=optional")

	envs = append(envs, "LOG_FORMAT=json")

	for k, v := range c.ExtraEnv {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}

	return envs
}

func (c *Config) Run(ctx context.Context, repo string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, c.Command, repo)
	cmd.Env = c.ToEnv()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run renovate: %w\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	return stdout.Bytes(), nil
}
