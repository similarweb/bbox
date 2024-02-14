/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bbox/teamcity"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// triggerCmd represents the trigger command
var triggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		tcc := teamcity.NewTeamCityClient(teamcityURL, teamcityUsername, teamcityPassword)
		err := tcc.TriggerBuild("ClientServer_Sandbox_Deployments_Multi_Deploy")
		if err != nil {
			log.Error("Error:", err)
		}

	},
}

func init() {
	rootCmd.AddCommand(triggerCmd)
}
