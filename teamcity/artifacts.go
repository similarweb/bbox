package teamcity

import (
	"bbox/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// BuildHasArtifact returns true if the build has artifacts
func (tcc *Client) BuildHasArtifact(buildId int) bool {
	artifactChildren, _ := tcc.GetArtifactChildren(buildId)
	return artifactChildren.Count > 0
}

// GetArtifactChildren returns the children of an artifact if any
func (tcc *Client) GetArtifactChildren(buildID int) (ArtifactChildrenResponse, error) {
	getUrl := fmt.Sprintf("%s/httpAuth/app/rest/builds/id:%d/%s", tcc.baseUrl, buildID, "artifacts/children/")
	log.Debug("getting build children from: ", getUrl)

	req, err := http.NewRequest("GET", getUrl, nil)

	if err != nil {
		return ArtifactChildrenResponse{}, err
	}
	req.SetBasicAuth(tcc.username, tcc.password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := tcc.client.Do(req)
	if err != nil {
		return ArtifactChildrenResponse{}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error closing response body: %s", err)
		}
	}(resp.Body)

	var ArtifactChildren ArtifactChildrenResponse
	err = json.NewDecoder(resp.Body).Decode(&ArtifactChildren)

	if err != nil {
		return ArtifactChildrenResponse{}, err
	}

	return ArtifactChildren, nil
}

// GetArtifactContentByPath GetArtifactContent returns the content of an artifact
func (tcc *Client) GetArtifactContentByPath(path string) ([]byte, error) {
	getUrl := fmt.Sprintf("%s%s", tcc.baseUrl, path)
	log.Debug("getting artifact content from: ", getUrl)

	req, err := http.NewRequest("GET", getUrl, nil)

	if err != nil {
		return []byte{}, err
	}
	req.SetBasicAuth(tcc.username, tcc.password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := tcc.client.Do(req)
	if err != nil {
		return nil, err
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

// GetAllBuildTypeArtifacts returns all artifacts from a buildID and buildTypeId as a zip file
func (tcc *Client) GetAllBuildTypeArtifacts(buildID int, buildTypeId string) ([]byte, error) {
	getUrl := fmt.Sprintf("%s/downloadArtifacts.html?buildId=%d&buildTypeId=%s", tcc.baseUrl, buildID, buildTypeId)

	log.Debug("Getting all artifacts from: ", getUrl)

	req, err := http.NewRequest("GET", getUrl, nil)

	if err != nil {
		return []byte{}, err
	}
	req.SetBasicAuth(tcc.username, tcc.password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := tcc.client.Do(req)
	if err != nil {
		return nil, err
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

// DownloadArtifacts downloads all artifacts to given path and unzips them
func (tcc *Client) DownloadArtifacts(buildID int, buildTypeId string, destPath string) error {
	content, err := tcc.GetAllBuildTypeArtifacts(buildID, buildTypeId)
	if err != nil {
		log.Errorf("error getting artifacts content: %s", err)
		return err
	}

	// if size of content is 0, then no artifacts were found
	if len(content) == 0 {
		return errors.New("artifacts not found")
	}

	err = utils.CreateDir(destPath)
	if err != nil {
		log.Errorf("error creating dir %s: %s", destPath, err)
		return err
	}
	// create uuid for temporary artifacts zip file, to prevent overwriting
	fileID := uuid.New().String()
	artifactsZip := filepath.Join(destPath, fileID+"-artifacts.zip")

	log.WithField("artifactsPath", destPath).Debug("writing Artifacts to path")
	err = utils.WriteContentToFile(artifactsZip, content)
	if err != nil {
		log.Errorf("error writing content to file: %s", err)
		return err
	}

	err = utils.UnzipFile(artifactsZip, destPath)
	if err != nil {
		log.Errorf("error unzipping artifacts: %s", err)
		return err
	}

	err = os.Remove(artifactsZip)
	if err != nil {
		log.Errorf("error deleteing zip: %s", err)
		return err
	}
	return nil
}
