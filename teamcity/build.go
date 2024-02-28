package teamcity

import (
	"bbox/pkg/display"
	"bbox/pkg/types"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	log "github.com/sirupsen/logrus"
)

type BuildService service

// GetBuildStatus returns the status of a build
func (bs *BuildService) GetBuildStatus(buildId int) (types.BuildStatusResponse, error) {
	getUrl := fmt.Sprintf("%s/id:%d", "app/rest/builds", buildId)

	req, err := bs.client.NewRequestWrapper("GET", getUrl, nil)

	if err != nil {
		return types.BuildStatusResponse{}, errors.Wrapf(err, "error getting build status for build id: %d", buildId)
	}

	resp, err := bs.client.client.Do(req)
	if err != nil {
		return types.BuildStatusResponse{}, errors.Wrapf(err, "error getting build status for build id: %d", buildId)
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
		return types.BuildStatusResponse{}, errors.Wrap(err, "error reading response body")
	}

	return *bsr, nil
}

// TriggerBuild triggers a build with parameters
func (bs *BuildService) TriggerBuild(buildTypeId string, branchName string, params map[string]string) (types.TriggerBuildWithParametersResponse, error) {
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
		return types.TriggerBuildWithParametersResponse{}, errors.Wrapf(err, "error creating request to trigger build")
	}

	resp, err := bs.client.client.Do(req)
	if err != nil {
		log.Errorf("error executing request to trigger build: %v", err)
		return types.TriggerBuildWithParametersResponse{}, errors.Wrapf(err, "error executing request to trigger build")
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

// TriggerBuilds triggers the builds for each set of build parameters, wait and download artifacts if needed using work group
func (c *Client) TriggerBuilds(params []types.BuildParameters, waitForBuilds bool, waitTimeout time.Duration, multiArtifactsPath string) {
	buildFailed := false
	resultsChan := make(chan types.BuildResult)

	var wg sync.WaitGroup

	for _, param := range params {
		// Increment the WaitGroup's counter for each goroutine
		wg.Add(1)

		// Launch a goroutine for each set of parameters
		go func(p types.BuildParameters) {
			defer wg.Done() // Decrement the counter when the goroutine completes

			log.WithFields(log.Fields{
				"branchName":        p.BranchName,
				"buildTypeId":       p.BuildTypeId,
				"properties":        p.PropertiesFlag,
				"downloadArtifacts": p.DownloadArtifacts,
			}).Debug("triggering Build")

			triggerResponse, err := c.Build.TriggerBuild(p.BuildTypeId, p.BranchName, p.PropertiesFlag)

			if err != nil {
				log.Error("error triggering build: ", err)
				return
			}

			log.WithFields(log.Fields{
				"buildName": triggerResponse.BuildType.Name,
				"webURL":    triggerResponse.WebURL,
			}).Info("Build Triggered")

			downloadedArtifacts := false
			status := "UNKNOWN"

			if waitForBuilds {
				log.Infof("waiting for build %s", triggerResponse.BuildType.Name)

				build, err := c.Build.WaitForBuild(triggerResponse.BuildType.Name, triggerResponse.ID, waitTimeout)

				if err != nil {
					log.Errorf("error waiting for build %s: %s", triggerResponse.BuildType.Name, err.Error())
				}

				log.WithFields(log.Fields{
					"buildStatus": build.Status,
					"buildState":  build.State,
				}).Infof("build %s Finished", triggerResponse.BuildType.Name)

				if p.DownloadArtifacts && err == nil && c.Artifacts.BuildHasArtifact(build.ID) {
					log.Infof("downloading Artifacts for %s", triggerResponse.BuildType.Name)

					err = c.Artifacts.DownloadAndUnzipArtifacts(build.ID, p.BuildTypeId, multiArtifactsPath)
					if err != nil {
						log.Errorf("error downloading artifacts for build %s: %s", triggerResponse.BuildType.Name, err.Error())
					}

					downloadedArtifacts = err == nil
				}

				status = build.Status

				if status != "SUCCESS" {
					buildFailed = true
				}
			}

			resultsChan <- types.BuildResult{
				BuildName:           triggerResponse.BuildType.Name,
				WebURL:              triggerResponse.WebURL,
				BranchName:          p.BranchName,
				BuildStatus:         status,
				DownloadedArtifacts: downloadedArtifacts,
				Error:               err,
			}
		}(param)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var results []types.BuildResult
	for result := range resultsChan {
		results = append(results, result)
	}

	display.ResultsTable(results)

	if buildFailed {
		log.Error("one or more builds failed, more info in links above")
		os.Exit(2)
	}
}

// WaitForBuild waits for a build to finish
func (bs *BuildService) WaitForBuild(buildName string, buildNumber int, timeout time.Duration) (types.BuildStatusResponse, error) {
	var status types.BuildStatusResponse

	baseDelay := 5 * time.Second // Initial delay of 5 seconds
	maxDelay := 20 * time.Second // Maximum delay

	var factor uint = 2 // Factor by which the delay is multiplied each attempt

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

			log.Infof("build %s has not finished yet, rechecking in %d seconds", buildName, time.Duration(delay.Seconds()))

			return delay
		}),
	)

	return status, nil
}
