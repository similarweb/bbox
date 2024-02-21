package cmd

import (
	"bbox/teamcity"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"regexp"
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
	downloadArtifacts bool
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
	multiTriggerCmd.PersistentFlags().StringSliceVarP(&buildParamsCombinations, "build-params-combination", "c", []string{}, "Combinations as 'buildTypeID;branchName;downloadArtifactsBool;key1=value1&key2=value2' format. Repeatable. example: 'byBuildId;master;true;key=value&key2=value2'")
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

		if isValidBuildID(parts[0]) == false {
			return nil, fmt.Errorf("invalid buildTypeID: %s", parts[0])
		}

		if isValidBranchName(parts[1]) == false {
			return nil, fmt.Errorf("invalid branchName: %s", parts[1])
		}

		if parts[2] != "true" && parts[2] != "false" {
			return nil, fmt.Errorf("invalid downloadArtifacts boolean: %s", parts[2])
		}

		downloadArtifacts, valid := isValidDownloadArtifacts(parts[2])

		if valid == false {
			return nil, fmt.Errorf("invalid downloadArtifacts boolean: %s", parts[2])
		}

		properties, err := parseProperties(parts[3])

		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse properties: %s", parts[3])
		}

		parsed = append(parsed, BuildParameters{
			buildTypeId:       parts[0],
			branchName:        parts[1],
			downloadArtifacts: downloadArtifacts,
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

			if p.downloadArtifacts && client.BuildHasArtifact(build.ID) {
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

func isValidBuildID(buildID string) bool {
	matched, _ := regexp.MatchString("^[a-zA-Z0-9-_]+$", buildID)
	return matched
}

func isValidBranchName(branchName string) bool {
	matched, _ := regexp.MatchString("^[^ ~^:\\\\?*[\\]@{}/][^ ~^:\\\\?*[\\]@{}]*$", branchName)
	return matched
}

// isValidDownloadArtifacts checks if the downloadArtifacts string is either "true" or "false", case-insensitively.
// Returns a boolean indicating if the value is "true" or "false" and a second boolean validity flag.
func isValidDownloadArtifacts(downloadArtifacts string) (bool, bool) {
	normalized := strings.ToLower(downloadArtifacts)
	if normalized == "true" {
		return true, true
	} else if normalized == "false" {
		return false, true
	}

	return false, false
}

// parseProperties parses the properties from the command line and returns a map of string to string
func parseProperties(properties string) (map[string]string, error) {
	propertiesMap := make(map[string]string)
	for _, prop := range strings.Split(properties, "&") {
		kv := strings.SplitN(prop, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid property format: %s", prop)
		}
		key, value := kv[0], kv[1]
		if !validateParamKey(key) {
			return nil, fmt.Errorf("invalid property key: %s", key)
		}
		if !validateParamValue(value) {
			return nil, fmt.Errorf("invalid property value: %s", value)
		}
		propertiesMap[key] = value
	}
	return propertiesMap, nil
}

// validateParamKey checks if the parameter is valid key and returns a boolean
func validateParamKey(key string) bool {
	matched, _ := regexp.MatchString(`^\w+[a-zA-Z0-9\\;,*/_.-]*`, key)
	return matched
}

// validateParamValue checks if the parameter is valid value and returns a boolean
func validateParamValue(value string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\\;,*/@:_.-]*$`, value)
	return matched
}
