package clean

import (
	"bbox/teamcity"
	"fmt"
	"net/url"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var vcsRootCmdName string = "vcsRoot"

var vcsRootCmd = &cobra.Command{
	Use:   vcsRootCmdName,
	Short: "Print project names of unused VCS roots",
	Run: func(cmd *cobra.Command, args []string) {
		teamcityUsername, _ := cmd.Root().PersistentFlags().GetString("teamcity-username")
		teamcityPassword, _ := cmd.Root().PersistentFlags().GetString("teamcity-password")
		teamcityURL, _ := cmd.Root().PersistentFlags().GetString("teamcity-url")
		startMow := time.Now()
		url, err := url.Parse(teamcityURL)
		if err != nil {
			log.Errorf("error parsing TeamCity URL: %s", err)
			os.Exit(2)
		}

		client := teamcity.NewTeamCityClient(url, teamcityUsername, teamcityPassword)
		logger := log.WithField("teamcityURL", url.String())

		logger.Info("going to print the TeamCity vcsRoot that unused.")

		err = client.VCSRoot.GetUnusedVCSRoots()
		if err != nil {
			fmt.Printf("Error getting VCS roots: %s\n", err)
			return
		}
		fmt.Printf("Time taken: %s\n", time.Since(startMow))
	},
}

func init() {
	Cmd.AddCommand(vcsRootCmd)
}
