package bruteforce

import (
	"github.com/spf13/cobra"
)

func NewBruteforceCmd() *cobra.Command {
	return bruteforceCmd
}

var bruteforceCmd = &cobra.Command{
	Use:   "bruteforce",
	Short: "Bruteforce passwords management",
	Long:  `Bruteforce passwords management.`,
}
