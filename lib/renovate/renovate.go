package renovate

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Command  string            `toml:"command"`
	Platform string            `toml:"platform"`
	Token    string            `toml:"token"`
	Endpoint string            `toml:"endpoint"`
	LogLevel string            `toml:"log_level"`
	DryRun   bool              `toml:"dry_run"`
	Webhook  string            `toml:"webhook"`
	ExtraEnv map[string]string `toml:"extra_env"`
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
	if c.LogLevel != "" {
		envs = append(envs, fmt.Sprintf("LOG_LEVEL=%s", c.LogLevel))
	}
	if c.DryRun {
		envs = append(envs, "RENOVATE_DRY_RUN=true")
	}
	if c.Webhook != "" {
		envs = append(envs, fmt.Sprintf("RENOVATE_WEBHOOK=%s", c.Webhook))
	}

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
		return nil, fmt.Errorf("failed to run renovate: %w\nstderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}
