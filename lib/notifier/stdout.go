package notifier

import (
	"context"
	"fmt"
)

type StdoutNotifier struct{}

func NewStdoutNotifier() *StdoutNotifier {
	return &StdoutNotifier{}
}

func (n *StdoutNotifier) Notify(ctx context.Context, repo string, result string) error {
	fmt.Printf("Notification for %s:\n%s\n", repo, result)
	return nil
}
