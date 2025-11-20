package notifier

import (
	"context"
	"fmt"

	"github.com/snowmerak/renovates/lib/renovate"
)

type StdoutNotifier struct{}

func NewStdoutNotifier() *StdoutNotifier {
	return &StdoutNotifier{}
}

func (n *StdoutNotifier) Notify(ctx context.Context, repo string, updates []renovate.UpdateInfo) error {
	if len(updates) == 0 {
		fmt.Printf("Notification for %s:\nNo updates needed.\n", repo)
		return nil
	}

	fmt.Printf("Notification for %s:\nDependency Updates:\n", repo)
	for _, u := range updates {
		msg := fmt.Sprintf("- %s: %s -> %s", u.DepName, u.CurrentVersion, u.NewVersion)
		if u.PackageFile != "" {
			msg += fmt.Sprintf(" (%s)", u.PackageFile)
		}
		if u.UpdateType != "" {
			msg += fmt.Sprintf(" [%s]", u.UpdateType)
		}
		fmt.Println(msg)
	}
	return nil
}
