package utils

import (
	"os"
	"path/filepath"
)

// WriteContentToFile writes the provided content to a file at the specified path.
func WriteContentToFile(filePath string, content []byte) error {
	return os.WriteFile(filePath, content, 0644) // 0644: User can read/write, others can read
}

// CreateFileWithPath creates all necessary directories for the given path,
// and then creates the file if it does not exist.
func CreateFileWithPath(filePath string) error {
	// Extract the directory path
	dirPath := filepath.Dir(filePath)

	// Create all directories in the path, if necessary
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return err
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}
