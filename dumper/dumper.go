package dumper

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/GiorgiMakharadze/mongo-dumper-golang/config"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Dumper struct {
	MongoURL  string
	DumpDir   string
	AWSRegion string
	S3Bucket  string
	S3Client  *s3.Client
	Uploader  *manager.Uploader
}

func NewDumper(cfg *config.Config) *Dumper {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.AWSRegion),
	)
	if err != nil {
		log.Fatalf("unable to load AWS SDK config, %v", err)
	}

	s3Client := s3.NewFromConfig(awsCfg)

	uploader := manager.NewUploader(s3Client, func(u *manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024
		u.Concurrency = 10             
	})

	return &Dumper{
		MongoURL:  cfg.MongoURL,
		DumpDir:   cfg.DumpDir,
		AWSRegion: cfg.AWSRegion,
		S3Bucket:  cfg.S3Bucket,
		S3Client:  s3Client,
		Uploader:  uploader,
	}
}


func (d *Dumper) Dump() {
	now := time.Now()

	dateDir := now.Format("2006-01-02")
	timeDir := now.Format("15:04:05")
	fullPath := filepath.Join(d.DumpDir, dateDir, timeDir)

	err := os.MkdirAll(fullPath, os.ModePerm)
	if err != nil {
		log.Printf("Failed to create directory %s: %v", fullPath, err)
		return
	}

	cmd := exec.Command("mongodump", "--uri", d.MongoURL, "--out", fullPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("mongodump failed: %v\nOutput: %s", err, string(output))
		return
	}

	logrus.Infof("Successfully created dump at %s", fullPath)

	archivePath, err := d.compressDump(fullPath)
	if err != nil {
		log.Printf("Failed to compress dump: %v", err)
		return
	}
	defer os.RemoveAll(fullPath)

	s3Key := filepath.Join(dateDir, timeDir+".tar.gz")
	err = d.uploadToS3(archivePath, s3Key)
	if err != nil {
		log.Printf("Failed to upload dump to S3: %v", err)
		return
	}

	logrus.Infof("Successfully uploaded dump to s3://%s/%s", d.S3Bucket, s3Key)

	os.Remove(archivePath)
}

func (d *Dumper) compressDump(sourceDir string) (string, error) {
	archiveName := sourceDir + ".tar.gz"
	cmd := exec.Command("tar", "-czf", archiveName, "-C", filepath.Dir(sourceDir), filepath.Base(sourceDir))
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to compress dump using tar and pigz: %w", err)
	}
	return archiveName, nil
}


func (d *Dumper) uploadToS3(filePath, key string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for upload: %w", err)
	}
	defer file.Close()

	_, err = d.Uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(d.S3Bucket),
		Key:    aws.String(key),
		Body:   file,
	})

	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return nil
}
