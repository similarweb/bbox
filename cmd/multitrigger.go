package cmd

import (
	"bbox/teamcity"
	"fmt"
	"os"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var buildParamsCombinations []string
var multiTriggerCmdName string = "multi-trigger"
var multiArtifactsPath string = "./"

// BuildParameters Definition to hold each combination
type BuildParameters struct {
	buildTypeId       string
	branchName        string
	downloadArtifacts string
	propertiesFlag    map[string]string
}

var multiTriggerCmd = &cobra.Command{
	Use:   multiTriggerCmdName,
	Short: "Multi-trigger a TeamCity Build",
	Long:  `"Multi-trigger a TeamCity Build",`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Multi-triggering builds , parsing possible combinations")
		allCombinations, err := parseCombinations(buildParamsCombinations)
		if err != nil {
			log.Errorf("Failed to parse combinations: %v", err)
			os.Exit(1)
		}
		log.WithField("combinations", allCombinations).Debug("Here are the possible combinations")
		triggerBuilds(allCombinations)
	},
}

func init() {
	rootCmd.AddCommand(multiTriggerCmd)

	// Register the flags for Trigger command
	multiTriggerCmd.PersistentFlags().StringSliceVarP(&buildParamsCombinations, "build-params-combination", "c", []string{}, "Combinations in 'buildTypeID;branchName;key1=value1,key2=value2' format. Repeatable.")
	multiTriggerCmd.PersistentFlags().StringVar(&multiArtifactsPath, "artifacts-path", multiArtifactsPath, "Path to download Artifacts to")
}

// parseCombinations parses the combinations from the command line and returns a slice of BuildParameters
func parseCombinations(combinations []string) ([]BuildParameters, error) {
	parsed := make([]BuildParameters, 0, len(combinations))

	for _, combo := range combinations {
		parts := strings.Split(combo, ";")
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid combination format: %s", combo)
		}
		properties := make(map[string]string)
		if parts[3] != "" {
			for _, prop := range strings.Split(parts[3], "&") {
				kv := strings.SplitN(prop, "=", 2)
				if len(kv) != 2 {
					return nil, fmt.Errorf("invalid property format: %s", prop)
				}
				properties[kv[0]] = kv[1]
			}
		}
		parsed = append(parsed, BuildParameters{
			buildTypeId:       parts[0],
			branchName:        parts[1],
			downloadArtifacts: parts[2],
			propertiesFlag:    properties,
		})
	}

	return parsed, nil
}

// triggerBuilds triggers the builds for each set of build parameters
func triggerBuilds(params []BuildParameters) {
	client := teamcity.NewTeamCityClient(teamcityURL, teamcityUsername, teamcityPassword)

	var wg sync.WaitGroup
	for _, param := range params {
		// Increment the WaitGroup's counter for each goroutine
		wg.Add(1)

		// Launch a goroutine for each set of parameters
		go func(p BuildParameters) {
			defer wg.Done() // Decrement the counter when the goroutine completes

			logger := log.WithFields(log.Fields{
				"teamcityURL":       teamcityURL,
				"branchName":        branchName,
				"buildTypeId":       p.buildTypeId,
				"properties":        p.propertiesFlag,
				"downloadArtifacts": p.downloadArtifacts,
			})

			logger.Info("Triggering build")

			build, err := client.TriggerAndWaitForBuild(p.buildTypeId, p.branchName, p.propertiesFlag)
			if err != nil {
				log.Error(err)
			}

			logger.WithFields(log.Fields{
				"buildStatus": build.Status,
				"buildState":  build.State,
			}).Info("Build Finished")

			// if build is not successful, exit with error
			if build.Status != "SUCCESS" {
				log.Error("Build did not finish successfully")
				os.Exit(2)
			}

			if p.downloadArtifacts == "true" && client.BuildHasArtifact(build.ID) {
				logger.Info("Downloading Artifacts")

				err := client.DownloadArtifacts(build.ID, p.buildTypeId, multiArtifactsPath)
				if err != nil {
					log.Errorf("Error getting artifacts content: %s", err)
					os.Exit(2)
				}
			}
		}(param) // Pass the current parameters to the goroutine
	}

	wg.Wait() // Wait for all goroutines to complete
}