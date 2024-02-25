package teamcity

import (
	"bbox/utils"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/avast/retry-go/v4"
	log "github.com/sirupsen/logrus"
)

// TeamCityClient struct
type TeamCityClient struct {
	base_url string
	username string
	password string
	client   *http.Client
}

// NewTeamCityClient init
func NewTeamCityClient(base_url, username, password string) *TeamCityClient {
	return &TeamCityClient{
		base_url: base_url,
		username: username,
		password: password,
		client:   &http.Client{},
	}
}

// GetBuildStatus returns the status of a build
func (tcc *TeamCityClient) GetBuildStatus(buildId int) (BuildStatusResponse, error) {
	getUrl := fmt.Sprintf("%s/%s/id:%d", tcc.base_url, "app/rest/builds", buildId)
	log.Debug("Getting build status from: ", getUrl)

	req, err := http.NewRequest("GET", getUrl, nil)

	if err != nil {
		return BuildStatusResponse{}, err
	}
	req.SetBasicAuth(tcc.username, tcc.password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := tcc.client.Do(req)
	if err != nil {
		return BuildStatusResponse{}, err
	}
	defer resp.Body.Close()

	bs := new(BuildStatusResponse)
	err = json.NewDecoder(resp.Body).Decode(bs)

	if err != nil {
		return BuildStatusResponse{}, err
	}

	return *bs, nil
}

// TriggerBuildWithParameters triggers a build with parameters
func (tcc *TeamCityClient) TriggerBuildWithParameters(buildTypeId string, branchName string, params map[string]string) (TriggerBuildWithParametersResponse, error) {
	// Build the request payload with supplied parameters
	properties := []map[string]string{}
	for name, value := range params {
		properties = append(properties, map[string]string{"name": name, "value": value})
	}

	jsonData, err := json.Marshal(map[string]interface{}{
		"branchName": branchName,
		"buildType": map[string]string{
			"id": buildTypeId,
		},
		"properties": map[string]interface{}{
			"property": properties,
		},
	})

	if err != nil {
		return TriggerBuildWithParametersResponse{}, err
	}

	log.Debug("Triggering build with parameters: ", string(jsonData))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/httpAuth/app/rest/buildQueue", tcc.base_url), bytes.NewBuffer(jsonData))

	if err != nil {
		return TriggerBuildWithParametersResponse{}, err
	}

	req.SetBasicAuth(tcc.username, tcc.password) // Assuming basic auth for simplicity
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := tcc.client.Do(req)
	if err != nil {
		return TriggerBuildWithParametersResponse{}, err
	}

	defer resp.Body.Close()

	var triggerBuildResponse TriggerBuildWithParametersResponse
	err = json.NewDecoder(resp.Body).Decode(&triggerBuildResponse)

	if err != nil {
		log.Error("error reading response body:", err)
		return TriggerBuildWithParametersResponse{}, nil
	}

	log.WithFields(log.Fields{
		"triggeredWebURL":  triggerBuildResponse.WebURL,
		"buildTypeName":    triggerBuildResponse.BuildType.Name,
		"buildTypeProject": triggerBuildResponse.BuildType.ProjectName,
	}).Debug("triggered Response")

	return triggerBuildResponse, nil
}

// BuildHasArtifact returns true if the build has artifacts
func (tcc *TeamCityClient) BuildHasArtifact(buildId int) bool {
	artifactChildren, _ := tcc.GetArtifactChildren(buildId)
	return artifactChildren.Count > 0
}

// TriggerBuild triggers a build
func (tcc *TeamCityClient) TriggerBuild(buildId string, branchName string, params map[string]string) (TriggerBuildWithParametersResponse, error) {
	return tcc.TriggerBuildWithParameters(buildId, branchName, params)
}

// WaitForBuild waits for a build to finish
func (tcc *TeamCityClient) WaitForBuild(buildName string, buildNumber int, timeout time.Duration) (BuildStatusResponse, error) {
	var status BuildStatusResponse

	baseDelay := 5 * time.Second // Initial delay of 5 seconds
	maxDelay := 20 * time.Second // Maximum delay
	var factor uint = 2          // Factor by which the delay is multiplied each attempt

	var err error
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = retry.Do(
		func() error {
			status, err = tcc.GetBuildStatus(buildNumber)
			if err != nil {
				log.Errorf("error getting build status: %s", err)
				return err
			} else if status.State != "finished" {
				log.Debugf("%s state is: %s", buildName, status.State)
				return fmt.Errorf("build status is not finished")
			}
			return nil
		},
		retry.Attempts(0),
		retry.Context(ctx),
		// retry only if build is not finished yet, exit for another error
		retry.RetryIf(func(err error) bool {
			return strings.Contains(err.Error(), "build status is not finished")
		}),
		retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
			delay := baseDelay * time.Duration(n*factor)
			if time.Duration(delay.Seconds()) > time.Duration(maxDelay.Seconds()) {
				delay = maxDelay
			}
			log.Infof("build %s not finished yet, rechecking in %d seconds", buildName, time.Duration(delay.Seconds()))
			return delay
		}),
	)
	return status, nil
}

// GetArtifactChildren returns the children of an artifact if any
func (tcc *TeamCityClient) GetArtifactChildren(buildID int) (ArtifactChildrenResponse, error) {
	getUrl := fmt.Sprintf("%s/httpAuth/app/rest/builds/id:%d/%s", tcc.base_url, buildID, "artifacts/children/")
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
	defer resp.Body.Close()

	var ArtifactChildren ArtifactChildrenResponse
	err = json.NewDecoder(resp.Body).Decode(&ArtifactChildren)

	if err != nil {
		return ArtifactChildrenResponse{}, err
	}

	return ArtifactChildren, nil
}

// GetArtifactContentByPath GetArtifactContent returns the content of an artifact
func (tcc *TeamCityClient) GetArtifactContentByPath(path string) ([]byte, error) {
	getUrl := fmt.Sprintf("%s%s", tcc.base_url, path)
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("statusCode: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// GetAllBuildTypeArtifacts returns all artifacts from a buildID and buildTypeId as a zip file
func (tcc *TeamCityClient) GetAllBuildTypeArtifacts(buildID int, buildTypeId string) ([]byte, error) {
	getUrl := fmt.Sprintf("%s/downloadArtifacts.html?buildId=%d&buildTypeId=%s", tcc.base_url, buildID, buildTypeId)

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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("statusCode: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// DownloadArtifacts downloads all artifacts to given path and unzips them
func (tcc *TeamCityClient) DownloadArtifacts(buildID int, buildTypeId string, destPath string) error {
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

// ClearTeamCityQueue cancels all queued builds in TeamCity using the REST API.
func (tcc *TeamCityClient) ClearTeamCityQueue() error {

	// Prepare the request URL and headers
	reqURL := fmt.Sprintf("%s/app/rest/buildQueue", tcc.base_url)

	log.WithField("reqURL", reqURL).Debug("Clearing TeamCity build queue")

	req, err := http.NewRequest("DELETE", reqURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.SetBasicAuth(tcc.username, tcc.password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Perform the HTTP DELETE request
	response, err := tcc.client.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request to clear the queue: %v", err)
	}
	defer response.Body.Close()

	// Check the response. Expecting HTTP Status No Content (204) or OK (200) as a success indicator.
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent {
		bodyBytes, _ := ioutil.ReadAll(response.Body) // Try reading response body for error message, if any
		return fmt.Errorf("failed to clear the queue, status code: %d, response: %s", response.StatusCode, string(bodyBytes))
	}

	return nil
}
