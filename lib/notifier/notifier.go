package notifier

import (
	"context"

	"github.com/snowmerak/renovates/lib/renovate"
)

type Notifier interface {
	Notify(ctx context.Context, repo string, updates []renovate.UpdateInfo) error
}
