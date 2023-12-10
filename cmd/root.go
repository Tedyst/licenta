package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/cmd/bruteforce"
	"github.com/tedyst/licenta/cmd/extract"
	"github.com/tedyst/licenta/cmd/migrate"
	"github.com/tedyst/licenta/cmd/nvd"
	"github.com/tedyst/licenta/cmd/scan"
	"github.com/tedyst/licenta/cmd/tasks"
	"github.com/tedyst/licenta/cmd/user"
	"github.com/tedyst/licenta/telemetry"
	"github.com/ttys3/slogx"
)

var rootCmd = &cobra.Command{
	Use:   "licenta",
	Short: "An app for verifying the security of your databases",
	Long:  `This app allows you to verify the security of your databases by checking for vulnerabilities and misconfigurations. It also allows you to extract username and passwords from Git repositories/Docker images and try them against your databases. It features a REST API, which should be used by the frontend.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		initConfig(cmd)
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		level := "info"
		if viper.GetBool("debug") {
			level = "debug"
		}
		handler := slogx.New(slogx.WithTracing(), slogx.WithLevel(level), slogx.WithFullSource(), slogx.WithFormat(viper.GetString("output")))
		slog.SetDefault(handler)

		if viper.GetBool("telemetry") {
			if err := telemetry.InitTelemetry(viper.GetString("telemetry-collector-endpoint")); err != nil {
				return err
			}
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if err := telemetry.ShutdownTelemetry(); err != nil {
			return err
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
	v := viper.New()
	v.SetEnvPrefix("")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	bindFlags(cmd, v)
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug mode")
	rootCmd.PersistentFlags().String("output", "cli", "Output format")
	rootCmd.PersistentFlags().Bool("telemetry", false, "Enable telemetry")
	rootCmd.PersistentFlags().String("telemetry-collector-endpoint", "", "Telemetry collector endpoint")

	rootCmd.AddCommand(user.NewUserCmd())
	rootCmd.AddCommand(extract.NewExtractCmd())
	rootCmd.AddCommand(scan.NewScanCmd())
	rootCmd.AddCommand(nvd.NewNvdCmd())
	rootCmd.AddCommand(bruteforce.NewBruteforceCmd())
	rootCmd.AddCommand(tasks.GetTasksCmd())
	rootCmd.AddCommand(migrate.NewMigrateCmd())
}
