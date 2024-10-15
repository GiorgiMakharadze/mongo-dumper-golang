package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURL string
	DumpDir  string
	Schedule string
}

func LoadConfig() *Config {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found. Proceeding with environment variables.")
	}

	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		log.Fatal("MONGO_URL environment variable not set")
	}

	dumpDir := os.Getenv("DUMP_DIR")
	if dumpDir == "" {
		// Default dump directory
		dumpDir = "./cyclix-dumps"
	}

	// Schedule in cron format. Every 30 minutes.
	schedule := "0 */30 * * * *"

	return &Config{
		MongoURL: mongoURL,
		DumpDir:  dumpDir,
		Schedule: schedule,
	}
}
