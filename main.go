package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

	for _, repo := range repos {
		fmt.Printf("Running renovate for %s...\n", repo)
		output, err := cfg.Run(context.Background(), repo)
		if err != nil {
			log.Printf("failed to run renovate for %s: %v", repo, err)
			continue
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
			if err := n.Notify(context.Background(), repo, updates); err != nil {
				log.Printf("failed to notify for %s: %v", repo, err)
			}
		}
	}
}
