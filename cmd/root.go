package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/cmd/bruteforce"
	"github.com/tedyst/licenta/cmd/extract"
	"github.com/tedyst/licenta/cmd/nvd"
	"github.com/tedyst/licenta/cmd/scan"
	"github.com/tedyst/licenta/cmd/user"
)

var rootCmd = &cobra.Command{
	Use:   "licenta",
	Short: "A template for building Go applications",
	Long:  `A template for building Go applications.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		initConfig(cmd)
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			panic(err)
		}
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		configName := strings.ReplaceAll(f.Name, "-", "")

		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

func initConfig(cmd *cobra.Command) {
	cmd.MarkFlagRequired("database")

	v := viper.New()
	v.SetEnvPrefix("")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	bindFlags(cmd, v)
}

func init() {
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug mode")

	rootCmd.AddCommand(user.NewUserCmd())
	rootCmd.AddCommand(extract.NewExtractCmd())
	rootCmd.AddCommand(scan.NewScanCmd())
	rootCmd.AddCommand(nvd.NewNvdCmd())
	rootCmd.AddCommand(bruteforce.NewBruteforceCmd())
}
