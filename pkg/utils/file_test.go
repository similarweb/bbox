package utils_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"bbox/pkg/utils"

	"github.com/stretchr/testify/assert"
)

func TestUnzipFile(t *testing.T) {
	zipFilePath := "./testfiles/Archive.zip"
	destDir := "/tmp"

	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "unzip_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	err = utils.UnzipFile(zipFilePath, destDir)
	assert.NoError(t, err)

	// Verify that the files are correctly unzipped
	expectedFiles := []string{
		"file1.txt",
		"file2.txt",
		"subdir/file3.txt",
	}

	for _, file := range expectedFiles {
		filePath := filepath.Join(destDir, file)
		_, err := os.Stat(filePath)
		assert.NoError(t, err)
	}
}
