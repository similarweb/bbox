/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"bbox/version"

	"github.com/spf13/cobra"
)

var describeVersion bool

func init() {
	versionCmd.Flags().BoolVarP(&describeVersion, "describe", "d", false, "Return full version description")
	RootCmd.AddCommand(versionCmd)
}

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of bbox",
	Run: func(cmd *cobra.Command, args []string) {
		if (describeVersion) {
			fmt.Println("bbox version:", version.GetFormattedVersion())
			return
		} else {
			fmt.Println("bbox version:", version.GetVersion())
		}
	},
}
