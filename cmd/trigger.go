package cmd

import (
	"bbox/teamcity"
	"net/url"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var branchName = "master"
var artifactsPath = "./"

var (
	buildTypeId         string
	propertiesFlag      map[string]string
	downloadArtifacts   bool
	waitForBuild        bool
	waitForBuildTimeout = 15 * time.Minute
)

var triggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "Trigger a single TeamCity Build",
	Long:  `Trigger a single TeamCity Build`,
	Run: func(cmd *cobra.Command, args []string) {

		url, err := url.Parse(teamcityURL)
		if err != nil {
			log.Errorf("error parsing TeamCity URL: %s", err)
			os.Exit(2)
		}
		client := teamcity.NewTeamCityClient(*url, teamcityUsername, teamcityPassword)

		log.WithFields(log.Fields{
			"teamcityURL":       teamcityURL,
			"branchName":        branchName,
			"buildTypeId":       buildTypeId,
			"properties":        propertiesFlag,
			"downloadArtifacts": downloadArtifacts,
			"artifactsPath":     artifactsPath,
		}).Debug("triggering Build")

		triggerResponse, err := client.Build.TriggerBuild(buildTypeId, branchName, propertiesFlag)

		if err != nil {
			log.Error("error triggering build: ", err)
			os.Exit(2)
		}

		log.WithFields(log.Fields{
			"buildName": triggerResponse.BuildType.Name,
			"webURL":    triggerResponse.WebURL,
		}).Info("build Triggered")

		downloadedArtifacts := false
		status := "UNKNOWN"

		if waitForBuilds {
			log.Infof("waiting for build %s", triggerResponse.BuildType.Name)

			build, err := client.Build.WaitForBuild(triggerResponse.BuildType.Name, triggerResponse.ID, waitForBuildTimeout)

			if err != nil {
				log.Error("error waiting for build: ", err)
				os.Exit(2)
			}

			log.WithFields(log.Fields{
				"buildStatus": build.Status,
				"buildState":  build.State,
			}).Infof("Build %s Finished", triggerResponse.BuildType.Name)

			if downloadArtifacts && err == nil && client.Artifacts.BuildHasArtifact(build.ID) {
				log.Infof("downloading Artifacts for %s", triggerResponse.BuildType.Name)
				err = client.Artifacts.DownloadAndUnzipArtifacts(build.ID, buildTypeId, artifactsPath)
				if err != nil {
					log.Errorf("error downloading artifacts for build %s: %s", triggerResponse.BuildType.Name, err.Error())
				}
				downloadedArtifacts = err == nil
			}

			status = build.Status
		}
		log.WithFields(log.Fields{
			"BuildName":           triggerResponse.BuildType.Name,
			"WebURL":              triggerResponse.BuildType.WebURL,
			"BranchName":          branchName,
			"BuildStatus":         status,
			"DownloadedArtifacts": downloadedArtifacts,
			"Error":               err,
		}).Info("Done triggering build")
	},
}

func init() {
	rootCmd.AddCommand(triggerCmd)

	triggerCmd.PersistentFlags().StringVarP(&buildTypeId, "build-type-id", "i", "", "The Build Type")
	triggerCmd.PersistentFlags().StringVar(&artifactsPath, "artifacts-path", artifactsPath, "Path to download Artifacts to")
	triggerCmd.PersistentFlags().BoolVarP(&waitForBuild, "wait-for-build", "w", waitForBuild, "Wait for build to finish and get status")
	triggerCmd.PersistentFlags().DurationVarP(&waitForBuildTimeout, "wait-timeout", "t", waitForBuildTimeout, "Timeout for waiting for build to finish")
	triggerCmd.PersistentFlags().BoolVarP(&downloadArtifacts, "download-artifacts", "d", downloadArtifacts, "Download Artifacts")
	triggerCmd.PersistentFlags().StringVarP(&branchName, "branch-name", "b", branchName, "The Branch Name")
	triggerCmd.PersistentFlags().StringToStringVarP(&propertiesFlag, "properties", "p", nil, "The properties in key=value format")
}
