package multitrigger

import (
	"bbox/pkg/types"
	"bbox/teamcity"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

// triggerBuilds triggers the builds for each set of build parameters, wait and download artifacts if needed using work group.
func triggerBuilds(c *teamcity.Client, parameters []types.BuildParameters, waitForBuilds bool, waitTimeout time.Duration, multiArtifactsPath string) {
	flowFailed := false
	resultsChan := make(chan types.BuildResult, len(parameters))

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
			}).Debug("triggering Build")

			triggerResponse, err := c.Build.TriggerBuild(p.BuildTypeID, p.BranchName, p.PropertiesFlag)
			if err != nil {
				log.Error("error triggering build: ", err)

				flowFailed = true

				resultsChan <- types.BuildResult{
					BuildName:           p.BuildTypeID,
					WebURL:              triggerResponse.WebURL,
					BranchName:          p.BranchName,
					BuildStatus:         "NOT_TRIGGERED",
					DownloadedArtifacts: false,
					Error:               errors.Wrap(err, "error triggering build"),
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

					resultsChan <- types.BuildResult{
						BuildName:           triggerResponse.BuildType.Name,
						WebURL:              triggerResponse.WebURL,
						BranchName:          p.BranchName,
						BuildStatus:         status,
						DownloadedArtifacts: false,
						Error:               errors.Wrap(err, "error waiting for build"),
					}

					return
				}

				log.WithFields(log.Fields{
					"buildStatus": build.Status,
					"buildState":  build.State,
				}).Infof("build %s Finished", triggerResponse.BuildType.Name)

				if p.DownloadArtifacts && err == nil && c.Artifacts.BuildHasArtifact(build.ID) {
					log.Infof("downloading Artifacts for %s", triggerResponse.BuildType.Name)

					err = c.Artifacts.DownloadAndUnzipArtifacts(build.ID, p.BuildTypeID, multiArtifactsPath)
					if err != nil {
						log.Errorf("error downloading artifacts for build %s: %s", triggerResponse.BuildType.Name, err.Error())

						flowFailed = true

						resultsChan <- types.BuildResult{
							BuildName:           triggerResponse.BuildType.Name,
							WebURL:              triggerResponse.WebURL,
							BranchName:          p.BranchName,
							BuildStatus:         build.Status,
							DownloadedArtifacts: false,
							Error:               errors.Wrap(err, "error downloading artifacts"),
						}

						return
					}

					downloadedArtifacts = err == nil
				}

				status = build.Status

				flowFailed = status != "SUCCESS"
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

	resultsTable(results)

	if flowFailed {
		log.Error("one or more builds failed, more info in table above")
		os.Exit(2)
	}
}