package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/CoinRock-Ventures/mongo-dumper-go/config"
	"github.com/CoinRock-Ventures/mongo-dumper-go/dumper"
	"github.com/CoinRock-Ventures/mongo-dumper-go/scheduler"
)

func main() {
	cfg := config.LoadConfig()

	err := os.MkdirAll(cfg.DumpDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create dump directory: %v", err)
	}

	d := dumper.NewDumper(cfg.MongoURL, cfg.DumpDir)

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
