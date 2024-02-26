package teamcity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/avast/retry-go/v4"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
)

// GetBuildStatus returns the status of a build
func (tcc *Client) GetBuildStatus(buildId int) (BuildStatusResponse, error) {
	getUrl := fmt.Sprintf("%s/%s/id:%d", tcc.baseUrl, "app/rest/builds", buildId)
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error closing response body: %s", err)
		}
	}(resp.Body)

	bs := new(BuildStatusResponse)
	err = json.NewDecoder(resp.Body).Decode(bs)

	if err != nil {
		return BuildStatusResponse{}, err
	}

	return *bs, nil
}

// TriggerBuildWithParameters triggers a build with parameters
func (tcc *Client) TriggerBuildWithParameters(buildTypeId string, branchName string, params map[string]string) (TriggerBuildWithParametersResponse, error) {
	// Build the request payload with supplied parameters
	var properties []map[string]string
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

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/httpAuth/app/rest/buildQueue", tcc.baseUrl), bytes.NewBuffer(jsonData))

	if err != nil {
		log.Errorf("error creating request: %v", err)
		return TriggerBuildWithParametersResponse{}, err
	}

	req.SetBasicAuth(tcc.username, tcc.password) // Assuming basic auth for simplicity
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := tcc.client.Do(req)
	if err != nil {
		log.Errorf("error executing request to trigger build: %v", err)
		return TriggerBuildWithParametersResponse{}, err
	}

	log.Debugf("response status code: %d", resp.StatusCode)
	log.Debugf("response status: %s", resp.Body)

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

// TriggerBuild triggers a build
func (tcc *Client) TriggerBuild(buildId string, branchName string, params map[string]string) (TriggerBuildWithParametersResponse, error) {
	return tcc.TriggerBuildWithParameters(buildId, branchName, params)
}

// WaitForBuild waits for a build to finish
func (tcc *Client) WaitForBuild(buildName string, buildNumber int, timeout time.Duration) (BuildStatusResponse, error) {
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
