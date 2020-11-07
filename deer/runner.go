package deer

import (
	"context"

	"github.com/jasonlvhit/gocron"
)

// Runner is responsible for scheduling jobs.
type Runner struct {
	cfg    *Config
	cronCh chan bool
	store  Store
}

// NewRunner creates runner instance.
func NewRunner(cfg *Config, store Store) *Runner {
	return &Runner{
		cfg:   cfg,
		store: store,
	}
}

// Start begins cron jobs.
func (r *Runner) Start(ctx context.Context) {
	for _, m := range r.cfg.Monitors {
		for _, s := range m.Services {
			for _, h := range s.HttpChecks {
				gocron.Every(h.IntervalSec).Seconds().Do(h.RunFn(r.store))
			}
		}
	}

	r.cronCh = gocron.Start()
	<-r.cronCh
}

func (r *Runner) Shutdown(ctx context.Context) {
	r.cronCh <- true
	gocron.Clear()
}
