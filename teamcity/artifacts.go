package teamcity

import (
	"bbox/utils"
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

type ArtifactsService service

type ArtifactChildren struct {
	Count int `json:"count"`
	File  []struct {
		Name             string `json:"name"`
		Size             int    `json:"size"`
		ModificationTime string `json:"modificationTime"`
		Href             string `json:"href"`
		Content          struct {
			Href string `json:"href"`
		} `json:"content"`
	} `json:"file"`
}

// BuildHasArtifact returns true if the build has artifacts
func (ac *ArtifactsService) BuildHasArtifact(buildId int) bool {
	artifactChildren, _ := ac.GetArtifactChildren(buildId)
	return artifactChildren.Count > 0
}

// GetArtifactChildren returns the children of an artifact if any
func (as *ArtifactsService) GetArtifactChildren(buildID int) (ArtifactChildren, error) {
	getUrl := fmt.Sprintf("httpAuth/app/rest/builds/id:%d/%s", buildID, "artifacts/children/")
	log.Debug("getting build children from: ", getUrl)

	req, err := as.client.NewRequestWrapper("GET", getUrl, nil)

	if err != nil {
		return ArtifactChildren{}, err
	}

	resp, err := as.client.client.Do(req)
	if err != nil {
		return ArtifactChildren{}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error closing response body: %s", err)
		}
	}(resp.Body)

	var artifactChildrens ArtifactChildren
	err = json.NewDecoder(resp.Body).Decode(&artifactChildrens)

	if err != nil {
		return ArtifactChildren{}, err
	}

	return artifactChildrens, nil
}

// GetArtifactContentByPath GetArtifactContent returns the content of an artifact
func (ac *ArtifactsService) GetArtifactContentByPath(path string) ([]byte, error) {

	req, err := ac.client.NewRequestWrapper("GET", path, nil)
	if err != nil {
		return []byte{}, err
	}

	resp, err := ac.client.client.Do(req)
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
func (ac *ArtifactsService) getAllBuildTypeArtifacts(buildID int, buildTypeId string) ([]byte, error) {
	getUrl := fmt.Sprintf("downloadArtifacts.html?buildId=%d&buildTypeId=%s", buildID, buildTypeId)
	req, err := ac.client.NewRequestWrapper("GET", getUrl, nil)

	if err != nil {
		return []byte{}, err
	}

	resp, err := ac.client.client.Do(req)
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
func (ac *ArtifactsService) DownloadArtifacts(buildID int, buildTypeId string, destPath string) error {
	content, err := ac.getAllBuildTypeArtifacts(buildID, buildTypeId)
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
