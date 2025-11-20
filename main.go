package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/snowmerak/renovates/lib/notifier"
	"github.com/snowmerak/renovates/lib/renovate"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: renovates <repo>")
		os.Exit(1)
	}
	repo := os.Args[1]

	cfg, err := renovate.LoadConfig("config.toml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	fmt.Printf("Running renovate for %s...\n", repo)
	output, err := cfg.Run(context.Background(), repo)
	if err != nil {
		log.Fatalf("failed to run renovate: %v", err)
	}

	message := renovate.ParseUpdates(output)

	var notifiers []notifier.Notifier
	for _, n := range cfg.Notifiers {
		switch n.Type {
		case "stdout":
			notifiers = append(notifiers, notifier.NewStdoutNotifier())
		case "webhook":
			notifiers = append(notifiers, notifier.NewWebhookNotifier(n.URL))
		}
	}

	for _, n := range notifiers {
		if err := n.Notify(context.Background(), repo, message); err != nil {
			log.Printf("failed to notify: %v", err)
		}
	}
}
