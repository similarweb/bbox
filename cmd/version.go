/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bbox/version"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of bbox",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("bbox version: v%s\n", version.GetVersion())
	},
}
