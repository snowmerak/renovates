package runner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/snowmerak/renovates/internal/models"
)

type Runner struct {
	RenovateCmd string
}

func NewRunner(cmd string) *Runner {
	return &Runner{RenovateCmd: cmd}
}

func (r *Runner) Run(repo models.Repository, outputChan chan<- string) error {
	cmd := exec.Command(r.RenovateCmd)
	cmd.Env = os.Environ()

	// Configure for single repo
	reposJSON, _ := json.Marshal([]string{repo.Name})
	cmd.Env = append(cmd.Env, fmt.Sprintf("RENOVATE_REPOSITORIES=%s", string(reposJSON)))
	cmd.Env = append(cmd.Env, "RENOVATE_AUTODISCOVER=false")
	// Ensure we don't write the discovery file again
	cmd.Env = append(cmd.Env, "RENOVATE_WRITE_DISCOVERED_REPOS=")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start renovate: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Stream stdout
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			outputChan <- fmt.Sprintf("[STDOUT] %s", scanner.Text())
		}
	}()

	// Stream stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			outputChan <- fmt.Sprintf("[STDERR] %s", scanner.Text())
		}
	}()

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("renovate execution failed: %w", err)
	}

	return nil
}
