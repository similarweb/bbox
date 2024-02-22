package cmd

import (
	"bbox/teamcity"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var clearQueueCmdName string = "clear-queue"

var clearQueueCmd = &cobra.Command{
	Use:   clearQueueCmdName,
	Short: "Clear the TeamCity Build Queue",
	Long:  `Clear the TeamCity Build Queue`,
	Run: func(cmd *cobra.Command, args []string) {

		teamcityClient := teamcity.NewTeamCityClient(teamcityURL, teamcityUsername, teamcityPassword)
		logger := log.WithField("teamcityURL", teamcityURL)

		logger.Info("going to clear the TeamCity queue.")

		err := teamcityClient.ClearTeamCityQueue()
		if err != nil {
			log.Error("error triggering build: ", err)
			os.Exit(2)
		}
		logger.Info("clearing the TeamCity queue was successful.")
	},
}

func init() {
	rootCmd.AddCommand(clearQueueCmd)
}
