// Package jobs runs recurring background work (cycle rollover, webhook
// delivery retries) as goroutines inside the main process. At self-hosted,
// small-team scale this avoids needing a separate queue + Redis.
package jobs

import (
	"context"
	"log/slog"
	"time"
)

type Job struct {
	Name     string
	Interval time.Duration
	Run      func(ctx context.Context) error
}

type Scheduler struct {
	jobs []Job
}

func NewScheduler() *Scheduler {
	return &Scheduler{}
}

func (s *Scheduler) Register(j Job) {
	s.jobs = append(s.jobs, j)
}

// Start launches one ticker goroutine per registered job. It returns
// immediately; jobs stop when ctx is cancelled.
func (s *Scheduler) Start(ctx context.Context) {
	for _, j := range s.jobs {
		go s.run(ctx, j)
	}
}

func (s *Scheduler) run(ctx context.Context, j Job) {
	ticker := time.NewTicker(j.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := j.Run(ctx); err != nil {
				slog.Error("job failed", "job", j.Name, "err", err)
			}
		}
	}
}
