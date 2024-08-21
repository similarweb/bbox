package version

import (
	"log"
	"os"
	"strings"
)

var (
	// Version is loaded from version.txt file
	version = loadVersionFromFile()
)

func GetVersion() string {
	return version
}

func loadVersionFromFile() string {
	data, err := os.ReadFile("version/version.txt")
	if err != nil {
		log.Fatalf("Failed to read version file: %v", err)
	}
	return strings.TrimSpace(string(data))
}
