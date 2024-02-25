package cmd

import (
	"bbox/teamcity"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var buildParamsCombinations []string
var multiTriggerCmdName = "multi-trigger"
var multiArtifactsPath = "./"
var waitForBuilds = true
var waitTimeout = 30 * time.Minute

// BuildParameters Definition to hold each combination
type BuildParameters struct {
	buildTypeId       string
	branchName        string
	downloadArtifacts bool
	propertiesFlag    map[string]string
}

type BuildResult struct {
	BuildName           string
	WebURL              string
	BranchName          string
	BuildStatus         string
	DownloadedArtifacts bool
	Error               error
}

var multiTriggerCmd = &cobra.Command{
	Use:   multiTriggerCmdName,
	Short: "Multi-trigger a TeamCity Build",
	Long:  `"Multi-trigger a TeamCity Build",`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("multi-triggering builds, parsing possible combinations")
		allCombinations, err := parseCombinations(buildParamsCombinations)
		if err != nil {
			log.Errorf("failed to parse combinations: %v", err)
			os.Exit(1)
		}
		log.WithField("combinations", allCombinations).Debug("Here are the possible combinations")

		triggerBuilds(allCombinations, waitForBuilds, waitTimeout)
	},
}

func init() {
	rootCmd.AddCommand(multiTriggerCmd)

	// Register the flags for Trigger command
	multiTriggerCmd.PersistentFlags().StringSliceVarP(&buildParamsCombinations, "build-params-combination", "c", []string{}, "Combinations as 'buildTypeID;branchName;downloadArtifactsBool;key1=value1&key2=value2' format. Repeatable. example: 'byBuildId;master;true;key=value&key2=value2'")
	multiTriggerCmd.PersistentFlags().StringVar(&multiArtifactsPath, "artifacts-path", multiArtifactsPath, "Path to download Artifacts to")
	multiTriggerCmd.PersistentFlags().BoolVarP(&waitForBuilds, "wait-for-builds", "w", waitForBuilds, "Wait for builds to finish and get status")
	multiTriggerCmd.PersistentFlags().DurationVarP(&waitTimeout, "wait-timeout", "t", waitTimeout, "Timeout for waiting for builds to finish")
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
func triggerBuilds(params []BuildParameters, waitForBuilds bool, waitTimeout time.Duration) {
	client := teamcity.NewTeamCityClient(teamcityURL, teamcityUsername, teamcityPassword)

	resultsChan := make(chan BuildResult)
	var wg sync.WaitGroup
	for _, param := range params {
		// Increment the WaitGroup's counter for each goroutine
		wg.Add(1)

		// Launch a goroutine for each set of parameters
		go func(p BuildParameters) {
			defer wg.Done() // Decrement the counter when the goroutine completes

			log.WithFields(log.Fields{
				"teamcityURL":       teamcityURL,
				"branchName":        branchName,
				"buildTypeId":       p.buildTypeId,
				"properties":        p.propertiesFlag,
				"downloadArtifacts": p.downloadArtifacts,
			}).Debug("triggering Build")

			triggerResponse, err := client.TriggerBuild(p.buildTypeId, p.branchName, p.propertiesFlag)

			if err != nil {
				log.Error("error triggering build: ", err)
				return
			}

			log.WithFields(log.Fields{
				"buildName": triggerResponse.BuildType.Name,
				"webURL":    triggerResponse.WebURL,
			}).Info("Build Triggered")

			downloadedArtifacts := false
			status := "UNKNOWN"

			if waitForBuilds {
				log.Infof("waiting for build %s", triggerResponse.BuildType.Name)

				build, err := client.WaitForBuild(triggerResponse.BuildType.Name, triggerResponse.ID, waitTimeout)

				if err != nil {
					log.Errorf("error waiting for build %s: %s", triggerResponse.BuildType.Name, err.Error())
				}

				log.WithFields(log.Fields{
					"buildStatus": build.Status,
					"buildState":  build.State,
				}).Infof("build %s Finished", triggerResponse.BuildType.Name)

				if p.downloadArtifacts && err == nil && client.BuildHasArtifact(build.ID) {
					log.Infof("downloading Artifacts for %s", triggerResponse.BuildType.Name)
					err = client.DownloadArtifacts(build.ID, p.buildTypeId, multiArtifactsPath)
					if err != nil {
						log.Errorf("error downloading artifacts for build %s: %s", triggerResponse.BuildType.Name, err.Error())
					}
					downloadedArtifacts = err == nil
				}
				status = build.Status
			}

			resultsChan <- BuildResult{
				BuildName:           triggerResponse.BuildType.Name,
				WebURL:              triggerResponse.WebURL,
				BranchName:          p.branchName,
				BuildStatus:         status,
				DownloadedArtifacts: downloadedArtifacts,
				Error:               err,
			}
		}(param)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	buildFailed := false

	var results []BuildResult
	for result := range resultsChan {
		results = append(results, result)
		if result.BuildStatus != "SUCCESS" {
			buildFailed = true
		}
	}

	displayResults(results)

	if buildFailed {
		log.Error("one or more builds failed, more info in links above")
		os.Exit(2)
	}
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

func displayResults(results []BuildResult) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	table.SetHeader([]string{"Build Name", "Branch Name", "Status", "Artifacts Downloaded", "Error", "Web URL"})
	table.SetHeaderColor(tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor}, tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor}, tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor}, tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor}, tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor}, tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor})
	table.SetBorders(tablewriter.Border{Left: false, Top: true, Right: false, Bottom: true})
	var data [][]string
	for _, result := range results {
		errorMessage := "None"
		if result.Error != nil {
			errorMessage = result.Error.Error()
		}
		data = append(data, [][]string{
			{result.BuildName, result.BranchName, result.BuildStatus, strconv.FormatBool(result.DownloadedArtifacts), errorMessage, fmt.Sprintf("\033]8;;%s\a%s\033]8;;\a", result.WebURL, "click_here")},
		}...)
	}

	for _, row := range data {
		status := row[2]
		statusColor := tablewriter.FgHiRedColor
		if status == "SUCCESS" {
			statusColor = tablewriter.FgHiGreenColor
		}
		// color row cells
		table.Rich(row, []tablewriter.Colors{{}, {}, {tablewriter.Bold, statusColor}, {}, {}, {tablewriter.Bold, tablewriter.FgHiBlueColor}})
	}

	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()
}
