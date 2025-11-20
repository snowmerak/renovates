package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

	fmt.Println("Output:")
	fmt.Println(string(output))
}
