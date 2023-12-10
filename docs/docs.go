package main

import (
	"github.com/spf13/cobra/doc"
	"github.com/tedyst/licenta/cmd"
)

func main() {
	rootCmd := cmd.GetRootCmd()

	err := doc.GenMarkdownTree(rootCmd, "./docs")
	if err != nil {
		panic(err)
	}
}
