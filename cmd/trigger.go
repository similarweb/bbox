package cmd

import (
	"bbox/teamcity"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var branchName string = "master"
var triggerCmdName string = "trigger"
var artifactsPath string = "./"

var (
	buildTypeId       string
	propertiesFlag    map[string]string
	downLoadArtifacts bool
)

var triggerCmd = &cobra.Command{
	Use:   triggerCmdName,
	Short: "Trigger a single TeamCity Build",
	Long:  `Trigger a single TeamCity Build`,
	Run: func(cmd *cobra.Command, args []string) {

		teamcityClient := teamcity.NewTeamCityClient(teamcityURL, teamcityUsername, teamcityPassword)

		logger := log.WithFields(log.Fields{
			"teamcityURL": teamcityURL,
			"branchName":  branchName,
			"buildTypeId": buildTypeId})

		logger.Info("Triggering Build")
		// todo - use trigger + wait for build
		build, err := teamcityClient.TriggerAndWaitForBuild(buildTypeId, branchName, propertiesFlag)
		if err != nil {
			log.Error("Error triggering build: ", err)
			os.Exit(2)
		}

		logger.WithFields(log.Fields{
			"buildStatus": build.Status,
			"buildState":  build.State,
			"buildID":     build.ID,
		}).Info("Build Finished")

		// if build is not successful, exit with error
		if build.Status != "SUCCESS" {
			log.Error("Build did not finish successfully")
			os.Exit(2)
		}

		if downLoadArtifacts && teamcityClient.BuildHasArtifact(build.ID) {
			logger.Info("Downloading Artifacts")

			err := teamcityClient.DownloadArtifacts(build.ID, buildTypeId, artifactsPath)
			if err != nil {
				log.Errorf("Error getting artifacts content: %s", err)
				os.Exit(2)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(triggerCmd)

	// Register the flags for Trigger command
	triggerCmd.PersistentFlags().StringVarP(&buildTypeId, "build-type-id", "i", "", "The Build Type")
	triggerCmd.PersistentFlags().BoolVar(&downLoadArtifacts, "download-artifacts", false, "Download Artifacts")
	triggerCmd.PersistentFlags().StringVar(&artifactsPath, "artifacts-path", artifactsPath, "Path to download Artifacts to")
	triggerCmd.PersistentFlags().StringVarP(&branchName, "branch-name", "b", branchName, "The Branch Name")
	triggerCmd.PersistentFlags().StringToStringVarP(&propertiesFlag, "properties", "p", nil, "The properties in key=value format")
}
