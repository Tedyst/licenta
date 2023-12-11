package tasks

import (
	"github.com/spf13/cobra"
)

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Run tasks from the command line",
	Long:  `This command allows you to run tasks from the command line. These commands are intended only for debugging failed tasks or for initial setup. For the normal use, use the web interface to schedule tasks onto the queue. These commands will update the database directly, so their status will be visible in the web interface.`,
}

func GetTasksCmd() *cobra.Command {
	return tasksCmd
}

func init() {
	tasksCmd.PersistentFlags().String("database", "", "Database connection string")
	tasksCmd.MarkFlagRequired("database")
}
