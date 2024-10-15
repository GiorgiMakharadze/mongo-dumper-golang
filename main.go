package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/GiorgiMakharadze/mongo-dumper-golang/config"
	"github.com/GiorgiMakharadze/mongo-dumper-golang/dumper"
	"github.com/GiorgiMakharadze/mongo-dumper-golang/scheduler"
)

func main() {
	cfg := config.LoadConfig()

	err := os.MkdirAll(cfg.DumpDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create dump directory: %v", err)
	}

	d := dumper.NewDumper(cfg)

	s := scheduler.NewScheduler()

	s.Start(cfg.Schedule, d.Dump)

	log.Println("MongoDB dump scheduler started.")

	d.Dump()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down scheduler...")
	s.Stop()
	log.Println("Scheduler stopped.")
}
