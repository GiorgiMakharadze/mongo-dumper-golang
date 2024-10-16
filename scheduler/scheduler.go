package scheduler

import (
	"context"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type Scheduler struct {
	cron    *cron.Cron
	jobLock sync.Mutex
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithSeconds()),
	}
}

func (s *Scheduler) Start(ctx context.Context, schedule string, job func(context.Context) error) error {
	_, err := s.cron.AddFunc(schedule, func() {
		logrus.Info("Scheduled job started.")
		if err := job(ctx); err != nil {
			logrus.Errorf("Scheduled job failed: %v", err)
		} else {
			logrus.Info("Scheduled job completed successfully.")
		}
	})
	if err != nil {
		return err
	}
	s.cron.Start()

	go func() {
		<-ctx.Done()
		logrus.Info("Context canceled. Stopping scheduler.")
		s.cron.Stop()
	}()

	return nil
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}
