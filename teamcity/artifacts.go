package teamcity

import (
	"bbox/pkg/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"bbox/pkg/utils"

	"github.com/pkg/errors"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type ArtifactsService service

// BuildHasArtifact returns true if the build has artifacts.
func (as *ArtifactsService) BuildHasArtifact(buildID int) bool {
	artifactChildren, err := as.GetArtifactChildren(buildID)
	if err != nil {
		log.Errorf("error getting artifact children: %s", err)
		return false
	}

	hasArtifacts := artifactChildren.Count > 0

	if !hasArtifacts {
		log.Debugf("$$$$$$$$ buildID: %d has artifacts: %t", buildID, hasArtifacts)
		log.Debugf("buildID: %d artifactChildren.Count: %d", buildID, artifactChildren.Count)
		log.Debugf("buildID: %d artifactChildren.File len: %d", buildID, len(artifactChildren.File))

		log.Debugf("buildID: %d sleeping and rechecking artifact children", buildID)
		time.Sleep(30 * time.Second)
		artifactChildren, err = as.GetArtifactChildren(buildID)
		hasArtifacts = artifactChildren.Count > 0

		log.Debugf("after sleeping buildID: %d artifactChildren.Count: %d", buildID, artifactChildren.Count)
		log.Debugf("after sleeping buildID: %d artifactChildren.File len: %d", buildID, len(artifactChildren.File))
	}

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
		return types.ArtifactChildren{}, errors.Wrap(err, "error getting artifact children")
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
		return types.ArtifactChildren{}, errors.Wrapf(err, "error decoding response body: %s", err)
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
		return nil, errors.Wrap(err, "error getting artifact content")
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

// getAllBuildTypeArtifacts returns all artifacts from a buildID and buildTypeId as a zip file.
func (as *ArtifactsService) getAllBuildTypeArtifacts(buildID int, buildTypeID string) ([]byte, error) {
	getURL := fmt.Sprintf("downloadArtifacts.html?buildId=%d&buildTypeId=%s", buildID, buildTypeID)

	req, err := as.client.NewRequestWrapper("GET", getURL, nil)
	if err != nil {
		return []byte{}, err
	}

	resp, err := as.client.client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting all artifacts for build id: %d", buildID)
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

// DownloadAndUnzipArtifacts downloads all artifacts to given path and unzips them.
func (as *ArtifactsService) DownloadAndUnzipArtifacts(buildID int, buildTypeID, destPath string) error {
	content, err := as.getAllBuildTypeArtifacts(buildID, buildTypeID)
	if err != nil {
		log.Errorf("error getting artifacts content: %s", err)
		return errors.Wrap(err, "error getting artifacts content")
	}

	// if size of content is 0, then no artifacts were found
	if len(content) == 0 {
		return errors.New("artifacts not found")
	}

	err = utils.CreateDir(destPath)
	if err != nil {
		log.Errorf("error creating dir %s: %s", destPath, err)
		return errors.Wrap(err, "error creating dir")
	}
	// create uuid for temporary artifacts zip file, to prevent overwriting
	fileID := uuid.New().String()
	artifactsZip := filepath.Join(destPath, fileID+"-artifacts.zip")

	log.WithField("artifactsPath", destPath).Debug("writing Artifacts to path")

	err = utils.WriteContentToFile(artifactsZip, content)
	if err != nil {
		log.Errorf("error writing content to file: %s", err)
		return errors.Wrap(err, "error writing content to file")
	}

	err = utils.UnzipFile(artifactsZip, destPath)
	if err != nil {
		log.Errorf("error unzipping artifacts: %s", err)
		return errors.Wrap(err, "error unzipping artifacts")
	}

	err = os.Remove(artifactsZip)
	if err != nil {
		log.Errorf("error deleting zip: %s", err)
		return errors.Wrap(err, "error deleting zip")
	}

	return nil
}
