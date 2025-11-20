package notifier

import "context"

type Notifier interface {
	Notify(ctx context.Context, repo string, result string) error
}
