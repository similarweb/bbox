package clean

import (
	"bbox/teamcity"
	"fmt"
	"net/url"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var vcsRootCmdName string = "vcsRoot"

var vcsRootCmd = &cobra.Command{
	Use:   vcsRootCmdName,
	Short: "Delete all unused VCS roots",
	Long:  `Delete all unused VCS roots. "Unused" VCS Root refers to a VCS Root that is neither linked to any build configurations nor included in any build templates.`,
	Run: func(cmd *cobra.Command, args []string) {
		teamcityUsername, _ := cmd.Root().PersistentFlags().GetString("teamcity-username")
		teamcityPassword, _ := cmd.Root().PersistentFlags().GetString("teamcity-password")
		teamcityURL, _ := cmd.Root().PersistentFlags().GetString("teamcity-url")

		url, err := url.Parse(teamcityURL)
		if err != nil {
			log.Errorf("error parsing TeamCity URL: %s", err)
			os.Exit(2)
		}

		client := teamcity.NewTeamCityClient(url, teamcityUsername, teamcityPassword)
		logger := log.WithField("teamcityURL", url.String())

		logger.Info("going to delete all unused TeamCity vcsRoot.")

		unusedVcsRoot, err := client.VCSRoot.GetUnusedVCSRoots()
		if err != nil {
			fmt.Printf("Error getting VCS roots: %s\n", err)
			return
		}
		logger.Infof(" %d vcs root have found", unusedVcsRoot)
	},
}

func init() {
	Cmd.AddCommand(vcsRootCmd)
}
