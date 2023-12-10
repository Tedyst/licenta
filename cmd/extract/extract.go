/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package extract

import (
	"github.com/spf13/cobra"
)

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Run the extract process",
	Long:  `This command allows you to run the extract process. The extract process will find all the passwords and usernames from a source and show them to you. It does not require a database running.`,
}

func NewExtractCmd() *cobra.Command {
	return extractCmd
}
