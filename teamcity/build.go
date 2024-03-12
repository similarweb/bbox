package teamcity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"bbox/pkg/types"

	"github.com/avast/retry-go/v4"
	log "github.com/sirupsen/logrus"
)

type BuildService service

// GetBuildStatus returns the status of a build.
func (bs *BuildService) GetBuildStatus(buildID int) (types.BuildStatusResponse, error) {
	getURL := fmt.Sprintf("%s/id:%d", "app/rest/builds", buildID)

	req, err := bs.client.NewRequestWrapper("GET", getURL, nil)
	if err != nil {
		return types.BuildStatusResponse{}, fmt.Errorf("error getting build status for buildID %d: %w", buildID, err)
	}

	resp, err := bs.client.client.Do(req)
	if err != nil {
		return types.BuildStatusResponse{}, fmt.Errorf("error getting build status for buildID %d: %w", buildID, err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error closing response body: %s", err)
		}
	}(resp.Body)

	bsr := new(types.BuildStatusResponse)

	err = json.NewDecoder(resp.Body).Decode(bsr)
	if err != nil {
		return types.BuildStatusResponse{}, fmt.Errorf("error reading response body: %w", err)
	}

	return *bsr, nil
}

// TriggerBuild triggers a build with parameters.
func (bs *BuildService) TriggerBuild(buildTypeID, branchName string, params map[string]string) (types.TriggerBuildWithParametersResponse, error) {
	// Build the request payload with supplied parameters
	properties := []map[string]string{}
	for name, value := range params {
		properties = append(properties, map[string]string{"name": name, "value": value})
	}

	data := map[string]interface{}{
		"branchName": branchName,
		"buildType": map[string]string{
			"id": buildTypeID,
		},
		"properties": map[string]interface{}{
			"property": properties,
		},
	}

	log.Debugf("Triggering build with parameters: %v ", data)

	req, err := bs.client.NewRequestWrapper("POST", "httpAuth/app/rest/buildQueue", data)
	if err != nil {
		log.Errorf("error creating request: %v", err)
		return types.TriggerBuildWithParametersResponse{}, fmt.Errorf("error creating request to trigger build: %w", err)
	}

	resp, err := bs.client.client.Do(req)
	if err != nil {
		log.Errorf("error executing request to trigger build: %v", err)
		return types.TriggerBuildWithParametersResponse{}, fmt.Errorf("error executing request to trigger build: %w", err)
	}

	log.Debugf("response status code: %d", resp.StatusCode)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error closing response body: %s", err)
		}
	}(resp.Body)

	var triggerBuildResponse types.TriggerBuildWithParametersResponse

	err = json.NewDecoder(resp.Body).Decode(&triggerBuildResponse)
	if err != nil {
		log.Error("error reading response body:", err)
		return types.TriggerBuildWithParametersResponse{}, nil
	}

	log.WithFields(log.Fields{
		"triggeredWebURL":  triggerBuildResponse.WebURL,
		"buildTypeName":    triggerBuildResponse.BuildType.Name,
		"buildTypeProject": triggerBuildResponse.BuildType.ProjectName,
	}).Debug("triggered Response")

	return triggerBuildResponse, nil
}

// WaitForBuild waits for a build to finish.
func (bs *BuildService) WaitForBuild(buildName string, buildNumber int, timeout time.Duration) (types.BuildStatusResponse, error) {
	var status types.BuildStatusResponse

	baseDelay := 5 * time.Second // Initial delay of 5 seconds
	maxDelay := 20 * time.Second // Maximum delay

	var factor uint = 2 // Factor by which the delay is multiplied each attempt

	var err error
	var errBuildNotFinished = errors.New("build status is not finished")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = retry.Do(
		func() error {
			status, err = bs.GetBuildStatus(buildNumber)

			if err != nil {
				log.Errorf("error getting build status: %s", err)
				return err
			}

			log.Debugf("%s state is: %s", buildName, status.State)

			if status.State != "finished" {
				return errBuildNotFinished
			}

			return nil
		},
		retry.Context(ctx),
		// retry only if build is not finished yet, exit for another error
		retry.RetryIf(func(err error) bool {
			return errors.Is(err, errBuildNotFinished)
		}),
		// default is 10, so we have to put 0 to disable it
		retry.Attempts(0),
		retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
			delay := baseDelay * time.Duration((n+1)*factor)
			if time.Duration(delay.Seconds()) > time.Duration(maxDelay.Seconds()) {
				delay = maxDelay
			}

			log.Infof("build %s has not finished yet, rechecking in %d seconds", buildName, time.Duration(delay.Seconds()))

			return delay
		}),
	)

	if err != nil && !errors.Is(err, errBuildNotFinished) {
		return status, fmt.Errorf("error waiting for build %s: %w", buildName, err)
	}

	return status, nil
}
