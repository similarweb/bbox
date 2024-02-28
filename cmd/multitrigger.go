package cmd

import (
	"bbox/pkg/display"
	"bbox/pkg/params"
	"bbox/pkg/types"
	"bbox/teamcity"
	"github.com/pkg/errors"
	"net/url"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var buildParamsCombinations []string
var multiTriggerCmdName = "multi-trigger"
var multiArtifactsPath = "./"
var waitForBuilds = true
var waitTimeout = 15 * time.Minute

var multiTriggerCmd = &cobra.Command{
	Use:   multiTriggerCmdName,
	Short: "Multi-trigger a TeamCity Build",
	Long:  `"Multi-trigger a TeamCity Build",`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("multi-triggering builds, parsing possible combinations")
		allCombinations, err := params.ParseCombinations(buildParamsCombinations)

		if err != nil {
			log.Errorf("failed to parse combinations: %v", err)
			os.Exit(1)
		}
		log.WithField("combinations", allCombinations).Debug("Here are the possible combinations")

		url, err := url.Parse(teamcityURL)
		if err != nil {
			log.Errorf("error parsing TeamCity URL: %s", err)
			os.Exit(2)
		}

		client := teamcity.NewTeamCityClient(url, teamcityUsername, teamcityPassword)

		triggerBuilds(client, allCombinations, waitForBuilds, waitTimeout, multiArtifactsPath)
	},
}

// triggerBuilds triggers the builds for each set of build parameters, wait and download artifacts if needed using work group
func triggerBuilds(c *teamcity.Client, params []types.BuildParameters, waitForBuilds bool, waitTimeout time.Duration, multiArtifactsPath string) {
	flowFailed := false
	resultsChan := make(chan types.BuildResult, len(params))

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
				flowFailed = true
				resultsChan <- types.BuildResult{
					BuildName:           p.BuildTypeId,
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

					err = c.Artifacts.DownloadAndUnzipArtifacts(build.ID, p.BuildTypeId, multiArtifactsPath)
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

	display.ResultsTable(results)

	if flowFailed {
		log.Error("one or more builds failed, more info in table above")
		os.Exit(2)
	}
}

func init() {
	rootCmd.AddCommand(multiTriggerCmd)

	// Register the flags for Trigger command
	multiTriggerCmd.PersistentFlags().StringSliceVarP(&buildParamsCombinations, "build-params-combination", "c", []string{}, "Combinations as 'buildTypeID;branchName;downloadArtifactsBool;key1=value1&key2=value2' format. Repeatable. example: 'byBuildId;master;true;key=value&key2=value2'")
	multiTriggerCmd.PersistentFlags().StringVar(&multiArtifactsPath, "artifacts-path", multiArtifactsPath, "Path to download Artifacts to")
	multiTriggerCmd.PersistentFlags().BoolVarP(&waitForBuilds, "wait-for-builds", "w", waitForBuilds, "Wait for builds to finish and get status")
	multiTriggerCmd.PersistentFlags().DurationVarP(&waitTimeout, "wait-timeout", "t", waitTimeout, "Timeout for waiting for builds to finish")
}
