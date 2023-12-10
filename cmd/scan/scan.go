package scan

import (
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan database type manually",
	Long:  `This command allows you to scan a database type manually. It does not update the database, so the results will not be visible in the web interface. The database connection string is not required, but is recommended for the bruteforce module. If a bruteforce is successful, the database will store that result in order to avoid repeating the bruteforce.`,
}

func NewScanCmd() *cobra.Command {
	return scanCmd
}
