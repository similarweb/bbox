package clean

import (
	"bbox/teamcity"
	"net/url"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var queueCmdName string = "queue"

var queueCmd = &cobra.Command{
	Use:   queueCmdName,
	Short: "Clear the TeamCity Build Queue",
	Run: func(cmd *cobra.Command, args []string) {
		teamcityUsername, _ := cmd.Root().PersistentFlags().GetString("teamcity-username")
		teamcityPassword, _ := cmd.Root().PersistentFlags().GetString("teamcity-password")
		teamcityURL, _ := cmd.Root().PersistentFlags().GetString("teamcity-url")

		url, err := url.Parse(teamcityURL)
		if err != nil {
			log.Errorf("error parsing TeamCity URL: %s", err)
			os.Exit(2)
		}

		client, err := teamcity.NewTeamCityClient(url, teamcityUsername, teamcityPassword)

		if err != nil {
			log.Errorf("error initializing TeamCity Client: %s", err)
			os.Exit(2)
		}

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
	Cmd.AddCommand(queueCmd)
}
