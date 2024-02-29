package clean

import (
	"github.com/spf13/cobra"
)

var clearCmdName string = "clean"

var CleanCmd = &cobra.Command{
	Use:   clearCmdName,
	Short: "Clean an unused or unwanted resources",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
