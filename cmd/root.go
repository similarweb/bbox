/*
Copyright © 2024 NAME HERE cregev
*/
package cmd

import (
	"bbox/logger"
	"os"

	"github.com/spf13/cobra"
)

var logLevel = "info"
var teamcityURL = "https://teamcity.similarweb.io/"

var (
	teamcityUsername string
	teamcityPassword string
)

// rootCmd represents the base command  called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bbox",
	Short: "bbox is a CLI tool for interacting with TeamCity and other CI/CD tools.",
	Long:  `bbox is a CLI tool for interacting with TeamCity and other CI/CD tools.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initCmd)
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", logLevel, "Log level (debug, info, warn, error, fatal, panic)")

	// TeamCity authentication
	rootCmd.PersistentFlags().StringVar(&teamcityUsername, "teamcity-username", "", "Teamcity username")
	rootCmd.PersistentFlags().StringVar(&teamcityPassword, "teamcity-password", "", "Teamcity password")
	rootCmd.PersistentFlags().StringVar(&teamcityURL, "teamcity-url", teamcityURL, "Teamcity URL")
	rootCmd.MarkFlagsRequiredTogether("teamcity-username", "teamcity-password")

}

func initCmd() {
	logger.InitializeLogger(logLevel)
}
