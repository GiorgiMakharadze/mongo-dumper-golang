package dumper

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/GiorgiMakharadze/mongo-dumper-golang/config"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cenkalti/backoff/v4"
)

type Dumper struct {
	MongoURL  string
	DumpDir   string
	AWSRegion string
	S3Bucket  string
	S3Client  *s3.Client
	Uploader  *manager.Uploader

	mu      sync.Mutex
	dumping bool
}

func NewDumper(cfg *config.Config) *Dumper {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.AWSRegion),
	)
	if err != nil {
		logrus.Fatalf("Unable to load AWS SDK config: %v", err)
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

func (d *Dumper) ValidateDependencies() error {
	requiredCommands := []string{"mongodump", "tar"}
	for _, cmd := range requiredCommands {
		if _, err := exec.LookPath(cmd); err != nil {
			return fmt.Errorf("required command %s not found in PATH", cmd)
		}
	}
	return nil
}

func (d *Dumper) Dump(ctx context.Context) error {
	d.mu.Lock()
	if d.dumping {
		logrus.Warn("Dump already in progress")
		d.mu.Unlock()
		return nil
	}
	d.dumping = true
	d.mu.Unlock()

	defer func() {
		d.mu.Lock()
		d.dumping = false
		d.mu.Unlock()
	}()

	now := time.Now()

	dateDir := now.Format("2006-01-02")
	timeDir := now.Format("150405") 
	fullPath := filepath.Join(d.DumpDir, dateDir, timeDir)

	err := os.MkdirAll(fullPath, os.ModePerm)
	if err != nil {
		logrus.Errorf("Failed to create directory %s: %v", fullPath, err)
		return err
	}

	cmd := exec.CommandContext(ctx, "mongodump", "--uri", d.MongoURL, "--out", fullPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("mongodump failed: %v, Output: %s", err, string(output))
		return err
	}

	logrus.Infof("Successfully created dump at %s", fullPath)

	archivePath, err := d.compressDump(fullPath)
	if err != nil {
		logrus.Errorf("Failed to compress dump: %v", err)
		return err
	}
	defer os.RemoveAll(fullPath)

	s3Key := filepath.Join(dateDir, timeDir+".tar.gz")
	uploadErr := d.uploadToS3WithRetry(ctx, archivePath, s3Key)
	if uploadErr != nil {
		logrus.Errorf("Failed to upload dump to S3 after retries: %v", uploadErr)
		return uploadErr
	}

	logrus.Infof("Successfully uploaded dump to s3://%s/%s", d.S3Bucket, s3Key)

	if err := os.Remove(archivePath); err != nil {
		logrus.Warnf("Failed to remove archive %s: %v", archivePath, err)
	}

	return nil
}

func (d *Dumper) compressDump(sourceDir string) (string, error) {
	archiveName := sourceDir + ".tar.gz"
	cmd := exec.Command("tar", "-czf", archiveName, "-C", filepath.Dir(sourceDir), filepath.Base(sourceDir))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to compress dump using tar: %w, Output: %s", err, string(output))
	}
	return archiveName, nil
}

func (d *Dumper) uploadToS3WithRetry(ctx context.Context, filePath, key string) error {
	operation := func() error {
		return d.uploadToS3(ctx, filePath, key)
	}

	backoffStrategy := backoff.NewExponentialBackOff()
	backoffStrategy.MaxElapsedTime = 2 * time.Minute

	err := backoff.Retry(operation, backoff.WithContext(backoffStrategy, ctx))
	if err != nil {
		return err
	}
	return nil
}

func (d *Dumper) uploadToS3(ctx context.Context, filePath, key string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for upload: %w", err)
	}
	defer file.Close()

	_, err = d.Uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(d.S3Bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return nil
}
