/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/api"
	database "github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/telemetry"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func serve() {
	telemetry.InitTelemetry()
	database.InitDatabase()
	app := api.InitializeFiber()
	app.Listen(":" + viper.GetString("port"))
}

func init() {
	serveCmd.Flags().Bool("sendgrid.enabled", false, "Enable Sendgrid")
	serveCmd.Flags().String("sendgrid.key", "", "Sendgrid API Key")
	serveCmd.MarkFlagsRequiredTogether("sendgrid.enabled", "sendgrid.key")

	serveCmd.Flags().String("email.sender", "no-reply@tedyst.ro", "Email sender")
	serveCmd.Flags().String("email.senderName", "Licenta", "Email sender name")

	serveCmd.Flags().String("baseurl", "http://localhost:8080", "Base URL")

	serveCmd.Flags().Bool("telemetry.metrics", false, "Enable metrics")
	serveCmd.Flags().Bool("telemetry.tracing", false, "Enable tracing")

	serveCmd.Flags().Int16P("port", "p", 5000, "Port to listen on")

	rootCmd.AddCommand(serveCmd)
}
