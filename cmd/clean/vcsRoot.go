package clean

import (
	"bbox/pkg/models"
	"bbox/teamcity"
	"net/url"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var vcsRootsCmdName string = "vcs"

var vcsRootsCmd = &cobra.Command{
	Use:   vcsRootsCmdName,
	Short: "Delete all unused vcs roots",
	Long:  `Delete all unused vcs roots. "Unused" vcs root refers to a vcs root that is neither linked to any build configurations nor included in any build templates.`,
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

		logger.Info("fetching all TeamCity vcsRoots.")
		allVcsRoots, err := client.VcsRoots.GetAllVcsRootsIDs()
		if err != nil {
			log.Error("error while trying to get all vcs roots: ", err)
			return
		}
		logger.Info("fetching all TeamCity projects")
		allTeamCityProjects, err := client.VcsRoots.GetAllProjects()
		if err != nil {
			log.Error("error while trying to get all projects: ", err)
			return
		}
		logger.Info("extracting all VCS roots templates from all projects")
		allVcsRootsTemplates, err := client.VcsRoots.GetAllVcsRootsTemplates(allTeamCityProjects)
		if err != nil {
			log.Error("error while trying to get all vcs roots templates: ", err)
			return
		}
		logger.Info("filtering all unused VCS roots")
		allUnusedVcsRoots, err := client.VcsRoots.GetUnusedVcsRootsIDs(allVcsRoots, allVcsRootsTemplates)
		if err != nil {
			log.Error("error while trying to get all unused vcs roots: ", err)
			return
		}

		model := models.ConfirmModel{}
		p := tea.NewProgram(model)
		activeModel, err := p.Run()
		if err != nil {
			log.Fatal("error while trying to start the program: ", err)
		}

		confirmedModel, ok := activeModel.(models.ConfirmModel)
		if !ok {
			log.Fatal("could not cast final model to ConfirmModel")
		}

		if confirmedModel.IsConfirmed() {
			logger.Info("deleting all unused VCS roots")
			numberOfDeletedVcsRoots, err := client.VcsRoots.DeleteUnusedVcsRoots(allUnusedVcsRoots)
			if err != nil {
				log.Errorf("Error while trying to delete unused VCS roots: %v", err)
				return
			}
			logger.Infof("%d unused VCS roots have been deleted.", numberOfDeletedVcsRoots)
		} else {
			logger.Info("deletion cancelled by the user.")
		}
	},
}

func init() {
	Cmd.AddCommand(vcsRootsCmd)
}
