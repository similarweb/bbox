package cmd

import (
	"bbox/pkg/params"
	"bbox/teamcity"
	"net/url"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var buildParamsCombinations []string
var multiTriggerCmdName = "multi-trigger"
var multiArtifactsPath = "./"
var waitForBuilds = true
var waitTimeout = 15 * time.Minute

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
		allCombinations, err := params.ParseCombinations(buildParamsCombinations)

		if err != nil {
			log.Errorf("failed to parse combinations: %v", err)
			os.Exit(1)
		}
		log.WithField("combinations", allCombinations).Debug("Here are the possible combinations")

		url, err := url.Parse(teamcityURL)
		if err != nil {
			log.Errorf("error parsing TeamCity URL: %s", err)
			os.Exit(2)
		}

		client := teamcity.NewTeamCityClient(*url, teamcityUsername, teamcityPassword)
		client.TriggerBuilds(allCombinations, waitForBuilds, waitTimeout, multiArtifactsPath)
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
