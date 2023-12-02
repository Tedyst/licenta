package tasks

import (
	"github.com/spf13/cobra"
)

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Run tasks",
	Long:  `Run tasks`,
}

func GetTasksCmd() *cobra.Command {
	return tasksCmd
}
