package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURL  string
	DumpDir   string
	Schedule  string
	AWSRegion string
	S3Bucket  string
}

func LoadConfig() *Config {
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
		dumpDir = "/tmp/cyclix-dumps"
	}

	schedule := "0 */30 * * * *"

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		log.Fatal("AWS_REGION environment variable not set")
	}

	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		log.Fatal("S3_BUCKET environment variable not set")
	}

	return &Config{
		MongoURL:  mongoURL,
		DumpDir:   dumpDir,
		Schedule:  schedule,
		AWSRegion: awsRegion,
		S3Bucket:  s3Bucket,
	}
}
