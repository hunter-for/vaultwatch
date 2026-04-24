package monitor

import (
	"context"
	"log"
	"time"
)

// Scheduler runs the monitor on a fixed interval until the context is cancelled.
type Scheduler struct {
	monitor  *Monitor
	interval time.Duration
}

// NewScheduler creates a Scheduler that ticks every interval.
func NewScheduler(m *Monitor, interval time.Duration) *Scheduler {
	if interval <= 0 {
		interval = 60 * time.Second
	}
	return &Scheduler{
		monitor:  m,
		interval: interval,
	}
}

// Run starts the polling loop. It performs an immediate check, then repeats
// every s.interval until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) error {
	log.Printf("[scheduler] starting — poll interval %s", s.interval)

	if err := s.tick(ctx); err != nil {
		log.Printf("[scheduler] initial check error: %v", err)
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[scheduler] stopping")
			return ctx.Err()
		case <-ticker.C:
			if err := s.tick(ctx); err != nil {
				log.Printf("[scheduler] check error: %v", err)
			}
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) error {
	return s.monitor.CheckAll(ctx)
}
