package deer

import (
	"sync"

	"github.com/jasonlvhit/gocron"
)

// Runner is responsible for scheduling jobs.
type Runner struct {
	cfg    *Config
	m      sync.Mutex
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
func (r *Runner) Start() chan bool {
	r.m.Lock()
	defer r.m.Unlock()

	for _, m := range r.cfg.Monitors {
		for _, s := range m.Services {
			for _, h := range s.HttpChecks {
				gocron.Every(h.IntervalSec).Seconds().Do(h.RunFn(r.store))
			}
		}
	}

	r.cronCh = gocron.Start()
	return r.cronCh
}

func (r *Runner) Shutdown() {
	r.m.Lock()
	defer r.m.Unlock()

	r.cronCh <- true
	gocron.Clear()
}
