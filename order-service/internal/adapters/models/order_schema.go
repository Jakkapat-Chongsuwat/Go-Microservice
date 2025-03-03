package models

import (
	"log"
	"os"
	"path/filepath"
)

var OrderEventSchema string

func LoadSchema(relativePath string) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %v", err)
	}
	fullPath := filepath.Join(wd, relativePath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		log.Fatalf("failed to load schema from %s: %v", fullPath, err)
	}
	OrderEventSchema = string(data)
}
