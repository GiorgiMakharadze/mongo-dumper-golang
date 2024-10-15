package scheduler

import (
	"log"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithSeconds()),
	}
}

func (s *Scheduler) Start(schedule string, job func()) {
	_, err := s.cron.AddFunc(schedule, job)
	if err != nil {
		log.Fatalf("Failed to add job to scheduler: %v", err)
	}
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}
