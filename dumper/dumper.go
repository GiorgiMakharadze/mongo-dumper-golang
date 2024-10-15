package dumper

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Dumper struct {
	MongoURL string
	DumpDir  string
}

func NewDumper(mongoURL, dumpDir string) *Dumper {
	return &Dumper{
		MongoURL: mongoURL,
		DumpDir:  dumpDir,
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
}
