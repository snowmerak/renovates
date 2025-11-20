package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"sync"

	"github.com/snowmerak/renovates/lib/discovery"
	"github.com/snowmerak/renovates/lib/notifier"
	"github.com/snowmerak/renovates/lib/renovate"
)

func main() {
	cfg, err := renovate.LoadConfig("config.toml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	var repos []string
	if len(os.Args) > 1 {
		repos = []string{os.Args[1]}
	} else if cfg.Discovery.Enabled {
		fmt.Println("Discovering repositories...")
		d, err := discovery.NewDiscoverer(cfg)
		if err != nil {
			log.Fatalf("failed to create discoverer: %v", err)
		}
		repos, err = d.ListRepositories(context.Background())
		if err != nil {
			log.Fatalf("failed to discover repositories: %v", err)
		}
		fmt.Printf("Found %d repositories: %v\n", len(repos), repos)
	} else {
		fmt.Println("Usage: renovates <repo> or enable discovery in config")
		os.Exit(1)
	}

	concurrency := cfg.Concurrency
	if concurrency < 1 {
		concurrency = 1
	}
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, repo := range repos {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(r string) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			fmt.Printf("Running renovate for %s...\n", r)
			output, err := cfg.Run(context.Background(), r)
			if err != nil {
				log.Printf("failed to run renovate for %s: %v", r, err)
				return
			}

			updates := renovate.ParseUpdates(output)

			var notifiers []notifier.Notifier
			for _, n := range cfg.Notifiers {
				switch n.Type {
				case "stdout":
					notifiers = append(notifiers, notifier.NewStdoutNotifier())
				case "webhook":
					notifiers = append(notifiers, notifier.NewWebhookNotifier(n.URL))
				case "teams":
					notifiers = append(notifiers, notifier.NewTeamsNotifier(n.URL))
				case "telegram":
					notifiers = append(notifiers, notifier.NewTelegramNotifier(n.Token, n.ChatID))
				}
			}

			for _, n := range notifiers {
				if err := n.Notify(context.Background(), r, updates); err != nil {
					log.Printf("failed to notify for %s: %v", r, err)
				}
			}
		}(repo)
	}

	wg.Wait()
}
