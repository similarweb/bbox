/*
Copyright © 2024 NAME HERE cregev
*/
package cmd

import (
	"os"

	"bbox/cmd/clean"
	"bbox/cmd/multitrigger"
	"bbox/logger"
	"github.com/spf13/cobra"
)

var (
	logLevel = "info"
)

var (
	TeamcityUsername string
	TeamcityPassword string
	TeamcityURL      string
)

// RootCmd represents the base command  called without any subcommands.
var RootCmd = &cobra.Command{
	Use:   "bbox",
	Short: "bbox is a CLI tool for interacting with TeamCity and other CI/CD tools.",
	Long:  `bbox is a CLI tool for interacting with TeamCity and other CI/CD tools.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initCmd)
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", logLevel, "Log level (debug, info, warn, error, fatal, panic)")

	// TeamCity authentication
	RootCmd.PersistentFlags().StringVar(&TeamcityUsername, "teamcity-username", "", "Teamcity username")
	RootCmd.PersistentFlags().StringVar(&TeamcityPassword, "teamcity-password", "", "Teamcity password")
	// get the TeamCity URL from environment variable if exists
	RootCmd.PersistentFlags().StringVar(&TeamcityURL, "teamcity-url", os.Getenv("BBOX_TEAMCITY_URL"), "Teamcity URL")
	RootCmd.MarkFlagsRequiredTogether("teamcity-username", "teamcity-password")
	RootCmd.AddCommand(clean.Cmd)
	RootCmd.AddCommand(multitrigger.Cmd)
}

func initCmd() {
	logger.InitializeLogger(logLevel)
}
