package multitrigger

import (
	"bbox/pkg/types"
	"bbox/teamcity"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

// triggerBuilds triggers the builds for each set of build parameters, wait and download artifacts if needed using work group.
func triggerBuilds(c *teamcity.Client, parameters []types.BuildParameters, waitForBuilds bool, waitTimeout time.Duration, multiArtifactsPath string, requireArtifacts bool) error {
	flowFailed := false
	resultsChan := make(chan types.BuildResult, len(parameters))
	errorChan := make(chan error, len(parameters))

	var wg sync.WaitGroup

	for _, param := range parameters {
		// Increment the WaitGroup's counter for each goroutine
		wg.Add(1)

		// Launch a goroutine for each set of parameters
		go func(p types.BuildParameters) {
			defer wg.Done() // Decrement the counter when the goroutine completes

			log.WithFields(log.Fields{
				"branchName":        p.BranchName,
				"buildTypeId":       p.BuildTypeID,
				"properties":        p.PropertiesFlag,
				"downloadArtifacts": p.DownloadArtifacts,
				"artifactsPath":     multiArtifactsPath,
				"requireArtifacts":  requireArtifacts,
				"waitForBuilds":     waitForBuilds,
			}).Debug("triggering Build")

			triggerResponse, err := c.Build.TriggerBuild(p.BuildTypeID, p.BranchName, p.PropertiesFlag)
			if err != nil {
				log.Error("error triggering build: ", err)

				flowFailed = true

				errorChan <- fmt.Errorf("error triggering build: %w", err)

				resultsChan <- types.BuildResult{
					BuildName:           p.BuildTypeID,
					WebURL:              triggerResponse.WebURL,
					BranchName:          p.BranchName,
					BuildStatus:         "NOT_TRIGGERED",
					DownloadedArtifacts: false,
					Error:               fmt.Errorf("error triggering build: %w", err),
				}

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

					flowFailed = true

					errorChan <- fmt.Errorf("error waiting for build: %w", err)

					resultsChan <- types.BuildResult{
						BuildName:           triggerResponse.BuildType.Name,
						WebURL:              triggerResponse.WebURL,
						BranchName:          p.BranchName,
						BuildStatus:         status,
						DownloadedArtifacts: false,
						Error:               fmt.Errorf("error waiting for build: %w", err),
					}

					return
				}

				status = build.Status

				log.WithFields(log.Fields{
					"buildStatus": build.Status,
					"buildState":  build.State,
				}).Infof("build %s finished", triggerResponse.BuildType.Name)

				if p.DownloadArtifacts && status == "SUCCESS" {
					downloadedArtifacts, err = handleArtifacts(c, build.ID, p.BuildTypeID, triggerResponse.BuildType.Name, multiArtifactsPath)
					if err != nil {
						log.Errorf("error handling artifacts for build %s: %s", triggerResponse.BuildType.Name, err.Error())

						flowFailed = true

						errorChan <- fmt.Errorf("error handling artifacts: %w", err)

						resultsChan <- types.BuildResult{
							BuildName:           triggerResponse.BuildType.Name,
							WebURL:              triggerResponse.WebURL,
							BranchName:          p.BranchName,
							BuildStatus:         status,
							DownloadedArtifacts: downloadedArtifacts,
							Error:               fmt.Errorf("error handling artifacts: %w", err),
						}

						return
					}
				}
			}
			// mark flow as failed if we had a build failure or error
			flowFailed = flowFailed || (status != "SUCCESS")

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

	wg.Wait()
	close(resultsChan)
	close(errorChan)

	var results []types.BuildResult
	for result := range resultsChan {
		results = append(results, result)
	}

	resultsTable(results)

	if flowFailed {
		log.Error("one or more builds failed, more info in table")
		// return the error from the channel
		return <-errorChan
	}

	log.Debugf("all builds finished successfully")

	return nil
}

// handleArtifacts handles the artifacts logic for a build, downloading and unzipping them if needed.
// Returns true if artifacts were downloaded, false otherwise.
func handleArtifacts(c *teamcity.Client, buildID int, buildTypeID, buildTypeName, artifactsPath string) (bool, error) {
	// if we have artifacts, download them
	if c.Artifacts.BuildHasArtifact(buildID) {
		log.Infof("downloading Artifacts for %s", buildTypeName)

		err := c.Artifacts.DownloadAndUnzipArtifacts(buildID, buildTypeID, artifactsPath)
		if err != nil {
			log.Errorf("error downloading artifacts for build %s: %s", buildTypeName, err.Error())
			return false, fmt.Errorf("error downloading artifacts: %w", err)
		}
		return true, nil
	}
	// if we require artifacts and did not get any, fail the build
	if requireArtifacts {
		log.Errorf("did not get artifacts for build %d, and requireArtifacts is true", buildID)
		return false, errors.New("build requires artifacts and did not produce any")
	}

	return false, nil
}
