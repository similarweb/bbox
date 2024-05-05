package teamcity

import (
	"bbox/pkg/types"
	"bbox/pkg/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type ArtifactsService struct {
	client *Client
}

// BuildHasArtifact returns true if the build has artifacts.
func (as *ArtifactsService) BuildHasArtifact(buildID int) bool {

	log.WithFields(log.Fields{
		"buildID": buildID,
	}).Debug("checking for artifacts")

	artifactChildren, err := as.GetArtifactChildren(buildID)

	if err != nil {
		log.WithFields(log.Fields{
			"buildID": buildID,
		}).Errorf("error getting artifact children: %s", err)
		return false
	}

	hasArtifacts := artifactChildren.Count > 0

	log.Debugf("buildID: %d has artifacts: %t", buildID, hasArtifacts)
	return hasArtifacts
}

// GetArtifactChildren returns the children of an artifact if any.
func (as *ArtifactsService) GetArtifactChildren(buildID int) (types.ArtifactChildren, error) {
	getURL := fmt.Sprintf("httpAuth/app/rest/builds/id:%d/%s", buildID, "artifacts/children/")
	log.Debug("getting build children from: ", getURL)

	req, err := as.client.NewRequestWrapper("GET", getURL, nil)
	if err != nil {
		return types.ArtifactChildren{}, err
	}

	resp, err := as.client.client.Do(req)

	if err != nil {
		return types.ArtifactChildren{}, fmt.Errorf("error getting artifact children: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error closing response body: %s", err)
		}
	}(resp.Body)

	var artifactChildren types.ArtifactChildren

	err = json.NewDecoder(resp.Body).Decode(&artifactChildren)
	if err != nil {
		return types.ArtifactChildren{}, fmt.Errorf("error decoding response body: %w", err)
	}

	// close
	err = resp.Body.Close()
	if err != nil {
		log.Errorf("error closing response body: %s", err)
	}

	return artifactChildren, nil
}

// GetArtifactContentByPath GetArtifactContent returns the content of an artifact.
func (as *ArtifactsService) GetArtifactContentByPath(path string) ([]byte, error) {
	req, err := as.client.NewRequestWrapper("GET", path, nil)
	if err != nil {
		return []byte{}, err
	}

	resp, err := as.client.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error getting artifact content: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error closing response body: %s", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("statusCode: %d", resp.StatusCode)
	}

	err = resp.Body.Close()
	if err != nil {
		log.Errorf("error closing response body: %s", err)
	}

	return io.ReadAll(resp.Body)
}

// GetAllBuildTypeArtifacts returns all artifacts from a buildID and buildTypeId as a zip file.
func (as *ArtifactsService) GetAllBuildTypeArtifacts(buildID int, buildTypeID string) ([]byte, error) {
	getURL := fmt.Sprintf("downloadArtifacts.html?buildId=%d&buildTypeId=%s", buildID, buildTypeID)

	req, err := as.client.NewRequestWrapper("GET", getURL, nil)
	if err != nil {
		return []byte{}, err
	}

	resp, err := as.client.client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("error getting all artifacts for buildID %d: %w", buildID, err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error closing response body: %s", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("statusCode: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// DownloadAndUnzipArtifacts downloads all artifacts  to given path and unzips them.
func (as *ArtifactsService) DownloadAndUnzipArtifacts(buildID int, buildTypeID, destPath string) error {
	content, err := as.GetAllBuildTypeArtifacts(buildID, buildTypeID)
	if err != nil {
		log.Errorf("error getting artifacts content: %s", err)
		return fmt.Errorf("error getting artifacts content: %w", err)
	}
	// if size of content is 0, then no artifacts were found
	if len(content) == 0 {
		return errors.New("artifacts not found")
	}

	err = utils.CreateDir(destPath)
	if err != nil {
		log.Errorf("error creating dir %s: %s", destPath, err)
		return fmt.Errorf("error creating dir: %w", err)
	}
	// create uuid for temporary artifacts zip file, to prevent overwriting
	fileID := uuid.New().String()
	artifactsZip := filepath.Join(destPath, fileID+"-artifacts.zip")

	log.WithField("artifactsPath", destPath).Debug("writing Artifacts to path")

	err = utils.WriteContentToFile(artifactsZip, content)
	if err != nil {
		log.Errorf("error writing content to file: %s", err)
		return fmt.Errorf("error writing content to file: %w", err)
	}

	err = utils.UnzipFile(artifactsZip, destPath)
	if err != nil {
		log.Errorf("error unzipping artifacts: %s", err)
		return fmt.Errorf("error unzipping artifacts: %w", err)
	}

	err = os.Remove(artifactsZip)
	if err != nil {
		log.Errorf("error deleting zip: %s", err)
		return fmt.Errorf("error deleting zip: %w", err)
	}

	return nil
}
