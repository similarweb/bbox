package multitrigger

import (
	"bbox/pkg/params"
	"bbox/teamcity"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"time"

	rootCmd "bbox/cmd"
)

var (
	buildParamsCombinations []string
	multiTriggerCmdName     = "multi-trigger"
	multiArtifactsPath      = "./"
	waitForBuilds           = true
	waitTimeout             = 15 * time.Minute
)

var Cmd = &cobra.Command{
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

		url, err := url.Parse(rootCmd.TeamcityURL)
		if err != nil {
			log.Errorf("error parsing TeamCity URL: %s", err)
			os.Exit(2)
		}

		client := teamcity.NewTeamCityClient(url, rootCmd.TeamcityUsername, rootCmd.TeamcityPassword)

		triggerBuilds(client, allCombinations, waitForBuilds, waitTimeout, multiArtifactsPath)
	},
}

func init() {
	Cmd.PersistentFlags().StringSliceVarP(&buildParamsCombinations, "build-params-combination", "c", []string{}, "Combinations as 'buildTypeID;branchName;downloadArtifactsBool;key1=value1&key2=value2' format. Repeatable. example: 'byBuildId;master;true;key=value&key2=value2'")
	Cmd.PersistentFlags().StringVar(&multiArtifactsPath, "artifacts-path", multiArtifactsPath, "Path to download Artifacts to")
	Cmd.PersistentFlags().BoolVarP(&waitForBuilds, "wait-for-builds", "w", waitForBuilds, "Wait for builds to finish and get status")
	Cmd.PersistentFlags().DurationVarP(&waitTimeout, "wait-timeout", "t", waitTimeout, "Timeout for waiting for builds to finish")
}
