/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"bbox/cmd"
	"bbox/cmd/multitrigger"
)

func main() {
	cmd.RootCmd.AddCommand(multitrigger.Cmd)
	cmd.Execute()
}
