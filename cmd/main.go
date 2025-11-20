package main

import (
	"fmt"
	"log"
	"os"

	"github.com/snowmerak/renovates/internal/discovery"
	"github.com/snowmerak/renovates/internal/models"
	"github.com/snowmerak/renovates/internal/notifier"
	"github.com/snowmerak/renovates/internal/runner"
	"github.com/snowmerak/renovates/internal/worker"
)

type ConsoleNotifier struct{}

func (n *ConsoleNotifier) Notify(result models.RenovateResult) error {
	status := "SUCCESS"
	if !result.Success {
		status = "FAILURE"
	}
	fmt.Printf(">>> NOTIFICATION: Repo: %s, Status: %s\n", result.Repository, status)
	return nil
}

func main() {
	renovateCmd := os.Getenv("RENOVATE_CMD")
	if renovateCmd == "" {
		renovateCmd = "renovate" // Default to renovate in PATH
	}

	fmt.Println("Starting Renovate Orchestrator...")
	fmt.Printf("Using Renovate command: %s\n", renovateCmd)

	// 1. Discovery
	fmt.Println("Discovering repositories...")
	discoverer := discovery.NewDiscoverer(renovateCmd)
	repos, err := discoverer.Discover()
	if err != nil {
		log.Printf("Discovery warning (might be empty if no repos found or error): %v", err)
		// For demo purposes, if discovery fails (e.g. because of local platform issues),
		// let's inject a dummy repo if none found, just to show the worker logic.
		if len(repos) == 0 {
			fmt.Println("No repos discovered (or error). Using dummy repo for demonstration.")
			repos = []models.Repository{{Name: "snowmerak/renovates"}}
		}
	}

	fmt.Printf("Discovered %d repositories.\n", len(repos))
	for _, r := range repos {
		fmt.Printf(" - %s\n", r.Name)
	}

	if len(repos) == 0 {
		fmt.Println("No repositories found. Exiting.")
		return
	}

	// 2. Worker Pool
	r := runner.NewRunner(renovateCmd)
	// Use ConsoleNotifier for demo, or WebhookNotifier if URL provided
	var n notifier.Notifier = &ConsoleNotifier{}

	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL != "" {
		n = notifier.NewWebhookNotifier(webhookURL)
	}

	pool := worker.NewPool(2, r, n) // 2 workers

	fmt.Println("Starting worker pool...")
	pool.Start(repos)
	fmt.Println("All jobs completed.")
}
