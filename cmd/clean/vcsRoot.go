package clean

import (
	"fmt"
	"net/url"
	"os"
	"sync"

	"bbox/pkg/models"
	"bbox/teamcity"

	"github.com/alitto/pond"
	tea "github.com/charmbracelet/bubbletea"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	PondWorkerPoolSize   = 50
	PondChannelTasksSize = 100
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
			os.Exit(1)
		}

		client, err := teamcity.NewTeamCityClient(url, teamcityUsername, teamcityPassword)
		if err != nil {
			log.Errorf("error creating TeamCity client: %s", err)
			os.Exit(1)
		}
		logger := log.WithField("teamcityURL", url.String())

		logger.Info("fetching all TeamCity VCS Roots.")
		allVcsRoots, err := client.VcsRoots.GetAllVcsRootsIDs()
		if err != nil {
			log.Error("error while trying to get all VCS Roots: ", err)
			os.Exit(1)
		}
		logger.Info("fetching all TeamCity projects")
		allTeamCityProjects, err := client.Project.GetAllProjects()
		if err != nil {
			log.Error("error while trying to get all projects: ", err)
			os.Exit(1)
		}

		logger.Info("extracting all VCS Roots templates from all projects")
		allVcsRootsTemplates, err := getAllVcsRootsTemplates(client, allTeamCityProjects)
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

		if len(allUnusedVcsRoots) == 0 {
			logger.Info("no unused VCS Roots found.")
			return

		}

		logger.Infof("%d unused VCS Roots found.", len(allUnusedVcsRoots))

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
				os.Exit(1)
			}

			confirmedModel, ok := activeModel.(models.ConfirmActionModel)
			if !ok {
				log.Error("could not cast final model to ConfirmModel")
				os.Exit(1)
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
				log.Info("deletion canceled by the user.")
			}
		}
	},
}

func init() {
	vcsRootsCmd.Flags().BoolVarP(&autoDelete, "confirm", "c", false, "Automatically confirm to delete all unused VCS Roots without prompting the user.")
	Cmd.AddCommand(vcsRootsCmd)
}

// GetAllVcsRootsTemplates fetches all VCS root templates for given TeamCity projects.
func getAllVcsRootsTemplates(client *teamcity.Client, allTeamCityProjects []string) ([]string, error) {
	pool := pond.New(PondWorkerPoolSize, PondChannelTasksSize)
	defer pool.StopAndWait()

	var mu sync.Mutex
	var errors []error

	allVcsRootsTemplates := []string{}

	for _, projectID := range allTeamCityProjects {
		localProjectID := projectID // capture range variable

		pool.Submit(func() {
			templateIDs, err := client.Project.GetProjectTemplates(localProjectID)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()

				return
			}

			templateVCSRootIDs, err := client.Template.GetVcsRootsIDsFromTemplates(templateIDs)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()

				return
			}

			mu.Lock()
			allVcsRootsTemplates = append(allVcsRootsTemplates, templateVCSRootIDs...)
			mu.Unlock()
		})
	}

	pool.StopAndWait()

	if len(errors) > 0 {
		return nil, fmt.Errorf("errors occurred while fetching templates: %v", errors)
	}

	return allVcsRootsTemplates, nil
}
