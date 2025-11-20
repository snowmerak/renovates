package notifier

import "github.com/snowmerak/renovates/internal/models"

type Notifier interface {
	Notify(result models.RenovateResult) error
}
