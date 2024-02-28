package display

import (
	"bbox/pkg/types"
	"github.com/olekukonko/tablewriter"
	"os"
	"strconv"
)

func ResultsTable(results []types.BuildResult) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	table.SetHeader([]string{"Build Name", "Branch Name", "Status", "Artifacts Downloaded", "Error", "Web URL"})
	table.SetHeaderColor(tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor}, tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor}, tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor}, tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor}, tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor}, tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor})
	table.SetBorders(tablewriter.Border{Left: false, Top: true, Right: false, Bottom: true})

	var data [][]string

	for _, result := range results {
		errorMessage := "None"
		if result.Error != nil {
			errorMessage = result.Error.Error()
		}

		data = append(data, [][]string{
			{result.BuildName, result.BranchName, result.BuildStatus, strconv.FormatBool(result.DownloadedArtifacts), errorMessage, result.WebURL},
		}...)
	}

	for _, row := range data {
		status := row[2]
		statusColor := tablewriter.FgHiRedColor

		if status == "SUCCESS" {
			statusColor = tablewriter.FgHiGreenColor
		}

		err := row[4]
		errorColor := tablewriter.FgHiRedColor

		if err == "None" {
			errorColor = tablewriter.FgWhiteColor
		}

		// color row cells
		table.Rich(row, []tablewriter.Colors{{}, {}, {tablewriter.Bold, statusColor}, {}, {tablewriter.Bold, errorColor}, {}})
	}

	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()
}
