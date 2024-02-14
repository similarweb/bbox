/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	log "bbox/logger"
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var ctx = context.Background()

var (
	teamcityUsername string
	teamcityPassword string
	teamcityURL      string
	logLevel         string
	Timeout          time.Duration
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bbox",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bbox.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "debug", "Log level (debug, info, warn, error, fatal, panic)")

	// TeamCity authentication
	rootCmd.PersistentFlags().StringVar(&teamcityUsername, "teamcity-username", "fargo", "Teamcity username")
	rootCmd.PersistentFlags().StringVar(&teamcityPassword, "teamcity-password", "", "Teamcity password")
	rootCmd.PersistentFlags().StringVar(&teamcityURL, "teamcity-url", "https://teamcity.similarweb.io", "Teamcity URL")
	rootCmd.MarkFlagsRequiredTogether("teamcity-username", "teamcity-password")

	log.InitializeLogger(ctx, logLevel)
}
