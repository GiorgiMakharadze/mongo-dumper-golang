package dumper

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/GiorgiMakharadze/mongo-dumper-golang/config"

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
	// Load AWS configuration
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.AWSRegion),
	)
	if err != nil {
		log.Fatalf("unable to load AWS SDK config, %v", err)
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(awsCfg)

	// Create uploader
	uploader := manager.NewUploader(s3Client)

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

	dateDir := now.Format("02-01")
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

	log.Printf("Successfully created dump at %s", fullPath)

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

	log.Printf("Successfully uploaded dump to s3://%s/%s", d.S3Bucket, s3Key)

	os.Remove(archivePath)
}

func (d *Dumper) compressDump(sourceDir string) (string, error) {
	archiveName := sourceDir + ".tar.gz"
	archiveFile, err := os.Create(archiveName)
	if err != nil {
		return "", fmt.Errorf("failed to create archive file: %w", err)
	}
	defer archiveFile.Close()

	gw := gzip.NewWriter(archiveFile)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	err = filepath.Walk(sourceDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(filepath.Dir(sourceDir), file)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(tw, f)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("error walking the path %q: %v", sourceDir, err)
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
