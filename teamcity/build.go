package teamcity

import (
	"bbox/pkg/display"
	"bbox/pkg/types"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	log "github.com/sirupsen/logrus"
)

type BuildService service

//type BuildStatusResponse struct {
//	ID        int    `json:"id"`
//	Status    string `json:"status"`
//	State     string `json:"state"`
//	Artifacts struct {
//		Href string `json:"href"`
//	} `json:"artifacts"`
//	SnapshotDependencies struct {
//		Count int `json:"count"`
//		Build []struct {
//			ID                  int    `json:"id"`
//			BuildTypeID         string `json:"buildTypeId"`
//			State               string `json:"state"`
//			BranchName          string `json:"branchName"`
//			Href                string `json:"href"`
//			WebURL              string `json:"webUrl"`
//			Customized          bool   `json:"customized"`
//			MatrixConfiguration struct {
//				Enabled bool `json:"enabled"`
//			} `json:"matrixConfiguration"`
//		} `json:"build"`
//	} `json:"snapshot-dependencies"`
//}
//
//type BuildResult struct {
//	BuildName           string
//	WebURL              string
//	BranchName          string
//	BuildStatus         string
//	DownloadedArtifacts bool
//	Error               error
//}
//
//// BuildParameters Definition to hold each combination
//type BuildParameters struct {
//	BuildTypeId       string
//	BranchName        string
//	DownloadArtifacts bool
//	PropertiesFlag    map[string]string
//}
//
//type TriggerBuildWithParametersResponse struct {
//	ID          int    `json:"id"`
//	BuildTypeID string `json:"buildTypeId"`
//	State       string `json:"state"`
//	Composite   bool   `json:"composite"`
//	Href        string `json:"href"`
//	WebURL      string `json:"webUrl"`
//	BuildType   struct {
//		ID          string `json:"id"`
//		Name        string `json:"name"`
//		Description string `json:"description"`
//		ProjectName string `json:"projectName"`
//		ProjectID   string `json:"projectId"`
//		Href        string `json:"href"`
//		WebURL      string `json:"webUrl"`
//	} `json:"buildType"`
//	WaitReason string `json:"waitReason"`
//	QueuedDate string `json:"queuedDate"`
//	Triggered  struct {
//		Type string `json:"type"`
//		Date string `json:"date"`
//		User struct {
//			Username string `json:"username"`
//			Name     string `json:"name"`
//			ID       int    `json:"id"`
//			Href     string `json:"href"`
//		} `json:"user"`
//	} `json:"triggered"`
//	SnapshotDependencies struct {
//		Count int `json:"count"`
//		Build []struct {
//			ID                  int    `json:"id"`
//			BuildTypeID         string `json:"buildTypeId"`
//			State               string `json:"state"`
//			BranchName          string `json:"branchName"`
//			DefaultBranch       bool   `json:"defaultBranch"`
//			Href                string `json:"href"`
//			WebURL              string `json:"webUrl"`
//			MatrixConfiguration struct {
//				Enabled bool `json:"enabled"`
//			} `json:"matrixConfiguration"`
//		} `json:"build"`
//	} `json:"snapshot-dependencies"`
//}

// GetBuildStatus returns the status of a build
func (bs *BuildService) GetBuildStatus(buildId int) (types.BuildStatusResponse, error) {
	getUrl := fmt.Sprintf("%s/id:%d", "app/rest/builds", buildId)

	req, err := bs.client.NewRequestWrapper("GET", getUrl, nil)

	if err != nil {
		return types.BuildStatusResponse{}, err
	}

	resp, err := bs.client.client.Do(req)
	if err != nil {
		return types.BuildStatusResponse{}, err
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
		return types.BuildStatusResponse{}, err
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
		return types.TriggerBuildWithParametersResponse{}, err
	}

	resp, err := bs.client.client.Do(req)
	if err != nil {
		log.Errorf("error executing request to trigger build: %v", err)
		return types.TriggerBuildWithParametersResponse{}, err
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
					err = c.Artifacts.DownloadArtifacts(build.ID, p.BuildTypeId, multiArtifactsPath)
					if err != nil {
						log.Errorf("error downloading artifacts for build %s: %s", triggerResponse.BuildType.Name, err.Error())
					}
					downloadedArtifacts = err == nil
				}
				status = build.Status
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

	buildFailed := false

	var results []types.BuildResult
	for result := range resultsChan {
		results = append(results, result)
		if result.BuildStatus != "SUCCESS" {
			buildFailed = true
		}
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
