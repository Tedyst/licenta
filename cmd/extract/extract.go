/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package extract

import (
	"github.com/spf13/cobra"
)

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

func NewExtractCmd() *cobra.Command {
	return extractCmd
}
