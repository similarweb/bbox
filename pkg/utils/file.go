package utils

import (
	"archive/zip"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
)

// WriteContentToFile writes the provided content to a file at the specified path.
func WriteContentToFile(filePath string, content []byte) error {
	return os.WriteFile(filePath, content, 0o644) // 0o644: User can read/write, others can read
}

// CreateDir creates all necessary directories for the given path
func CreateDir(path string) error {
	// Extract the directory path
	dirPath := filepath.Dir(path)
	// Create all directories in the path, if necessary
	return os.MkdirAll(dirPath, os.ModePerm)
}

func UnzipFile(zipFilePath, destDir string) error {
	log.Debugf("Unzipping file %s to %s", zipFilePath, destDir)

	r, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return errors.Wrap(err, "error opening zip file")
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			err := os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return errors.Wrap(err, "error creating dir")
			}
		} else {
			if err := CreateDir(fpath); err != nil {
				return errors.Wrap(err, "error creating dir")
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return errors.Wrap(err, "error opening file")
			}

			rc, err := f.Open()
			if err != nil {
				return errors.Wrap(err, "error opening file")
			}

			_, err = io.Copy(outFile, rc)

			// Close the file without defer to handle the error
			outFile.Close()
			rc.Close()

			if err != nil {
				return errors.Wrap(err, "error copying file")
			}
		}
	}

	return nil
}
