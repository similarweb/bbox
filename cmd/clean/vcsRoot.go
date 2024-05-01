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

var autoDelete bool

var vcsRootsCmdName string = "vcs"

var vcsRootsCmd = &cobra.Command{
	Use:   vcsRootsCmdName,
	Short: "Delete all unused VCS Roots",
	Long:  `Delete all unused VCS Roots. "Unused" VCS Root refers to a VCS Root that is neither linked to any build configurations nor included in any build templates.`,
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

		logger.Info("fetching all TeamCity VCS Roots.")
		allVcsRoots, err := client.VcsRoots.GetAllVcsRootsIDs()
		if err != nil {
			log.Error("error while trying to get all VCS Roots: ", err)
			os.Exit(1)
		}
		logger.Info("fetching all TeamCity projects")
		allTeamCityProjects, err := client.VcsRoots.GetAllProjects()
		if err != nil {
			log.Error("error while trying to get all projects: ", err)
			os.Exit(1)
		}
		logger.Info("extracting all VCS Roots templates from all projects")
		allVcsRootsTemplates, err := client.VcsRoots.GetAllVcsRootsTemplates(allTeamCityProjects)
		if err != nil {
			log.Error("error while trying to get all VCS Roots templates: ", err)
			os.Exit(1)
		}
		logger.Info("filtering all unused VCS Roots")
		allUnusedVcsRoots, err := client.VcsRoots.GetUnusedVcsRootsIDs(allVcsRoots, allVcsRootsTemplates)
		if err != nil {
			log.Error("error while trying to get all unused VCS Roots: ", err)
			os.Exit(1)
		}

		if autoDelete {
			logger.Info("deleting all unused VCS Roots")
			numberOfDeletedVcsRoots, err := client.VcsRoots.DeleteUnusedVcsRoots(allUnusedVcsRoots)
			if err != nil {
				log.Errorf("Error while trying to delete unused VCS Roots: %v", err)
				return
			}
			logger.Infof("%d unused VCS Roots have been deleted.", numberOfDeletedVcsRoots)
		} else {
			client.VcsRoots.PrintAllVcsRoots(allUnusedVcsRoots)
			model := models.NewConfirmActionModel()
			p := tea.NewProgram(model)
			activeModel, err := p.Run()
			if err != nil {
				log.Error("error while running confirmation model: ", err)
				os.Exit(2)
			}

			confirmedModel, ok := activeModel.(models.ConfirmActionModel)
			if !ok {
				log.Error("could not cast final model to ConfirmModel")
				os.Exit(2)
			}
			if confirmedModel.IsConfirmed() {

				logger.Info("deleting all unused VCS Roots")
				numberOfDeletedVCSRoots, err := client.VcsRoots.DeleteUnusedVcsRoots(allUnusedVcsRoots)
				if err != nil {
					log.Errorf("Error while trying to delete unused VCS Roots: %v", err)
					os.Exit(1)
				}
				logger.Infof("%d unused VCS Roots have been deleted.", numberOfDeletedVCSRoots)
			} else {
				log.Info("deletion cancelled by the user.")
			}
		}
		os.Exit(0)
	},
}

func init() {
	vcsRootsCmd.Flags().BoolVarP(&autoDelete, "confirm", "c", false, "Automatically confirm to delete all unused VCS Roots without prompting the user.")
	Cmd.AddCommand(vcsRootsCmd)
}
