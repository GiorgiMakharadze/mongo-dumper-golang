package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/GiorgiMakharadze/mongo-dumper-golang/config"
	"github.com/GiorgiMakharadze/mongo-dumper-golang/dumper"
	"github.com/GiorgiMakharadze/mongo-dumper-golang/scheduler"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	cfg := config.LoadConfig()

	err := os.MkdirAll(cfg.DumpDir, os.ModePerm)
	if err != nil {
		logrus.Fatalf("Failed to create dump directory: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := dumper.NewDumper(cfg)

	if err := d.ValidateDependencies(); err != nil {
		logrus.Fatalf("Dependency validation failed: %v", err)
	}

	s := scheduler.NewScheduler()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.Start(ctx, cfg.Schedule, d.Dump); err != nil {
			logrus.Errorf("Scheduler encountered an error: %v", err)
		}
	}()

	logrus.Info("MongoDB dump scheduler started.")

	go func() {
		if err := d.Dump(ctx); err != nil {
			logrus.Errorf("Initial dump failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down scheduler...")
	cancel()
	s.Stop()
	wg.Wait()
	logrus.Info("Scheduler stopped.")
}
