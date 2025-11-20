package worker

import (
	"fmt"
	"strings"
	"sync"

	"github.com/snowmerak/renovates/internal/models"
	"github.com/snowmerak/renovates/internal/notifier"
	"github.com/snowmerak/renovates/internal/runner"
)

type Pool struct {
	WorkerCount int
	Runner      *runner.Runner
	Notifier    notifier.Notifier
}

func NewPool(count int, runner *runner.Runner, notifier notifier.Notifier) *Pool {
	return &Pool{
		WorkerCount: count,
		Runner:      runner,
		Notifier:    notifier,
	}
}

func (p *Pool) Start(repos []models.Repository) {
	repoChan := make(chan models.Repository, len(repos))
	for _, r := range repos {
		repoChan <- r
	}
	close(repoChan)

	var wg sync.WaitGroup
	wg.Add(p.WorkerCount)

	for i := 0; i < p.WorkerCount; i++ {
		go func(workerID int) {
			defer wg.Done()
			for repo := range repoChan {
				p.processRepo(workerID, repo)
			}
		}(i)
	}

	wg.Wait()
}

func (p *Pool) processRepo(workerID int, repo models.Repository) {
	outputChan := make(chan string)
	var outputBuilder strings.Builder

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for line := range outputChan {
			outputBuilder.WriteString(line + "\n")
			// Here we could also stream to a real-time logger or websocket
			fmt.Printf("[Worker %d][%s] %s\n", workerID, repo.Name, line)
		}
	}()

	err := p.Runner.Run(repo, outputChan)
	close(outputChan)
	wg.Wait()

	result := models.RenovateResult{
		Repository: repo.Name,
		Success:    err == nil,
		Output:     outputBuilder.String(),
		Error:      err,
	}

	if p.Notifier != nil {
		if err := p.Notifier.Notify(result); err != nil {
			fmt.Printf("Failed to notify for %s: %v\n", repo.Name, err)
		}
	}
}
