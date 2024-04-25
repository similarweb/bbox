package multitrigger

import (
	"bbox/teamcity"
	"net/url"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	buildParamsCombinations []string
	multiTriggerCmdName     = "multi-trigger"
	multiArtifactsPath      = "./"
	waitForBuilds           = true
	waitTimeout             = 15 * time.Minute
	requireArtifacts        bool
)

var Cmd = &cobra.Command{
	Use:   multiTriggerCmdName,
	Short: "Multi-trigger a TeamCity Build",
	Long:  `"Multi-trigger a TeamCity Build",`,
	Run: func(cmd *cobra.Command, args []string) {
		teamcityUsername, _ := cmd.Root().PersistentFlags().GetString("teamcity-username")
		teamcityPassword, _ := cmd.Root().PersistentFlags().GetString("teamcity-password")
		teamcityURL, _ := cmd.Root().PersistentFlags().GetString("teamcity-url")
		log.Debug("multi-triggering builds, parsing possible combinations")
		allCombinations, err := parseCombinations(buildParamsCombinations)
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

		client := teamcity.NewTeamCityClient(url, teamcityUsername, teamcityPassword)

		err = triggerBuilds(client, allCombinations, waitForBuilds, waitTimeout, multiArtifactsPath, requireArtifacts)

		if err != nil {
			log.Errorf("trigger builds failed: %v", err)
			os.Exit(2)
		}
	},
}

func init() {
	Cmd.PersistentFlags().StringSliceVarP(&buildParamsCombinations, "build-params-combination", "c", []string{}, "Combinations as 'buildTypeID;branchName;downloadArtifactsBool;key1=value1&key2=value2' format. Repeatable. example: 'byBuildId;master;true;key=value&key2=value2'")
	Cmd.PersistentFlags().StringVar(&multiArtifactsPath, "artifacts-path", multiArtifactsPath, "Path to download Artifacts to")
	Cmd.PersistentFlags().BoolVarP(&waitForBuilds, "wait-for-builds", "w", waitForBuilds, "Wait for builds to finish and get status")
	Cmd.PersistentFlags().DurationVarP(&waitTimeout, "wait-timeout", "t", waitTimeout, "Timeout for waiting for builds to finish, default is 15 minutes")
	Cmd.PersistentFlags().BoolVar(&requireArtifacts, "require-artifacts", false, "If downloadArtifactsBool is true, and no artifacts found, return an error")
}
