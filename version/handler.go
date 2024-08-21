package version

import (
	"embed"
	"log"
	"strings"
)

var (
	// Version is loaded from version.txt file
	version = loadVersionFromFile()

	//go:embed version.txt
	embedFS embed.FS
)

func GetVersion() string {
	return version
}

func loadVersionFromFile() string {
	data, err := embedFS.ReadFile("version.txt")
	if err != nil {
		log.Fatalf("Failed to read version file: %v", err)
	}
	return strings.TrimSpace(string(data))
}
