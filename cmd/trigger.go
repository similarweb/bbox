package cmd

import (
	"bbox/teamcity"
	"bbox/utils"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var branchName string = "master"
var triggerCmdName string = "trigger"
var artifactsFile string = "./artifacts.zip"

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

		if downLoadArtifacts {
			logger.Info("Downloading Artifacts")

			content, err := teamcityClient.GetAllBuildTypeArtifacts(build.ID, buildTypeId)
			if err != nil {
				log.Errorf("Error getting artifacts content: %s", err)
				os.Exit(2)
			}

			logger.WithField("artifactsPath", artifactsFile).Info("Writing Artifacts to file")
			err = utils.WriteContentToFile(artifactsFile, content)
			if err != nil {
				log.Errorf("Error writing content to file: %s", err)
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
	triggerCmd.PersistentFlags().StringVar(&artifactsFile, "artifacts-file", artifactsFile, "Path & name for downloaded Artifacts file")
	triggerCmd.MarkFlagsRequiredTogether("download-artifacts", "artifacts-file")
	triggerCmd.PersistentFlags().StringVarP(&branchName, "branch-name", "b", branchName, "The Branch Name")
	triggerCmd.PersistentFlags().StringToStringVarP(&propertiesFlag, "properties", "p", nil, "The properties in key=value format")
}
