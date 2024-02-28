package teamcity

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	log "github.com/sirupsen/logrus"
)

type BuildService service

type BuildStatusResponse struct {
	ID        int    `json:"id"`
	Status    string `json:"status"`
	State     string `json:"state"`
	Artifacts struct {
		Href string `json:"href"`
	} `json:"artifacts"`
	SnapshotDependencies struct {
		Count int `json:"count"`
		Build []struct {
			ID                  int    `json:"id"`
			BuildTypeID         string `json:"buildTypeId"`
			State               string `json:"state"`
			BranchName          string `json:"branchName"`
			Href                string `json:"href"`
			WebURL              string `json:"webUrl"`
			Customized          bool   `json:"customized"`
			MatrixConfiguration struct {
				Enabled bool `json:"enabled"`
			} `json:"matrixConfiguration"`
		} `json:"build"`
	} `json:"snapshot-dependencies"`
}

type TriggerBuildWithParametersResponse struct {
	ID          int    `json:"id"`
	BuildTypeID string `json:"buildTypeId"`
	State       string `json:"state"`
	Composite   bool   `json:"composite"`
	Href        string `json:"href"`
	WebURL      string `json:"webUrl"`
	BuildType   struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		ProjectName string `json:"projectName"`
		ProjectID   string `json:"projectId"`
		Href        string `json:"href"`
		WebURL      string `json:"webUrl"`
	} `json:"buildType"`
	WaitReason string `json:"waitReason"`
	QueuedDate string `json:"queuedDate"`
	Triggered  struct {
		Type string `json:"type"`
		Date string `json:"date"`
		User struct {
			Username string `json:"username"`
			Name     string `json:"name"`
			ID       int    `json:"id"`
			Href     string `json:"href"`
		} `json:"user"`
	} `json:"triggered"`
	SnapshotDependencies struct {
		Count int `json:"count"`
		Build []struct {
			ID                  int    `json:"id"`
			BuildTypeID         string `json:"buildTypeId"`
			State               string `json:"state"`
			BranchName          string `json:"branchName"`
			DefaultBranch       bool   `json:"defaultBranch"`
			Href                string `json:"href"`
			WebURL              string `json:"webUrl"`
			MatrixConfiguration struct {
				Enabled bool `json:"enabled"`
			} `json:"matrixConfiguration"`
		} `json:"build"`
	} `json:"snapshot-dependencies"`
}

// GetBuildStatus returns the status of a build
func (bs *BuildService) GetBuildStatus(buildId int) (BuildStatusResponse, error) {
	getUrl := fmt.Sprintf("%s/id:%d", "app/rest/builds", buildId)

	req, err := bs.client.NewRequestWrapper("GET", getUrl, nil)

	if err != nil {
		return BuildStatusResponse{}, err
	}

	resp, err := bs.client.client.Do(req)
	if err != nil {
		return BuildStatusResponse{}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error closing response body: %s", err)
		}
	}(resp.Body)

	bsr := new(BuildStatusResponse)
	err = json.NewDecoder(resp.Body).Decode(bsr)

	if err != nil {
		return BuildStatusResponse{}, err
	}

	return *bsr, nil
}

// TriggerBuild triggers a build with parameters
func (bs *BuildService) TriggerBuild(buildTypeId string, branchName string, params map[string]string) (TriggerBuildWithParametersResponse, error) {
	// Build the request payload with supplied parameters
	properties := []map[string]string{}
	for name, value := range params {
		properties = append(properties, map[string]string{"name": name, "value": value})
	}

	data := map[string]interface{}{
		"branchName": branchName,
		"buildType": map[string]string{
			"id": buildTypeId,
		},
		"properties": map[string]interface{}{
			"property": properties,
		},
	}

	log.Debugf("Triggering build with parameters: %v ", data)

	req, err := bs.client.NewRequestWrapper("POST", "httpAuth/app/rest/buildQueue", data)

	if err != nil {
		log.Errorf("error creating request: %v", err)
		return TriggerBuildWithParametersResponse{}, err
	}

	resp, err := bs.client.client.Do(req)
	if err != nil {
		log.Errorf("error executing request to trigger build: %v", err)
		return TriggerBuildWithParametersResponse{}, err
	}

	log.Debugf("response status code: %d", resp.StatusCode)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error closing response body: %s", err)
		}
	}(resp.Body)

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

// WaitForBuild waits for a build to finish
func (bs *BuildService) WaitForBuild(buildName string, buildNumber int, timeout time.Duration) (BuildStatusResponse, error) {
	var status BuildStatusResponse

	baseDelay := 5 * time.Second // Initial delay of 5 seconds
	maxDelay := 20 * time.Second // Maximum delay
	var factor uint = 2          // Factor by which the delay is multiplied each attempt

	var err error
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = retry.Do(
		func() error {
			status, err = bs.GetBuildStatus(buildNumber)
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
