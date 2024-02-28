package cmd

import (
	"net/url"
	"os"

	"bbox/teamcity"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var clearQueueCmdName string = "clear-queue"

var clearQueueCmd = &cobra.Command{
	Use:   clearQueueCmdName,
	Short: "Clear the TeamCity Build Queue",
	Run: func(cmd *cobra.Command, args []string) {
		url, err := url.Parse(teamcityURL)
		if err != nil {
			log.Errorf("error parsing TeamCity URL: %s", err)
			os.Exit(2)
		}

		client := teamcity.NewTeamCityClient(url, teamcityUsername, teamcityPassword)
		logger := log.WithField("teamcityURL", url.String())

		logger.Info("going to clear the TeamCity queue.")

		err = client.Queue.ClearQueue()
		if err != nil {
			log.Error("error while trying to clear build queue: ", err)
			os.Exit(2)
		}
		logger.Info("clearing the TeamCity queue was successful.")
	},
}

func init() {
	rootCmd.AddCommand(clearQueueCmd)
}
