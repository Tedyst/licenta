package cmd

import (
	"log/slog"

	n "github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/bruteforce"
	database "github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/email"
	natsexchange "github.com/tedyst/licenta/messages/nats"
	"github.com/tedyst/licenta/tasks/local"
	"github.com/tedyst/licenta/tasks/nats"
)

var natsCmd = &cobra.Command{
	Use:   "nats",
	Short: "Run the NATS worker using the production configuration",
	Long:  `This command connects to the NATS server and listens to tasks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db := database.InitDatabase(viper.GetString("database"))

		natsConn, err := n.Connect(viper.GetString("nats"))
		if err != nil {
			return err
		}
		defer natsConn.Close()

		natsExchange, err := natsexchange.NewNATSExchange(natsConn)
		if err != nil {
			return err
		}

		bruteforceProvider := bruteforce.NewDatabaseBruteforceProvider(db)

		emailSender := email.NewSendGridEmailSender(
			viper.GetString("email-sendgrid"),
			viper.GetString("email-senderName"),
			viper.GetString("email-sender"),
		)

		localRunner := local.NewLocalRunner(viper.GetBool("debug"), emailSender, db, natsExchange, bruteforceProvider, viper.GetString("db-encryption-salt"))

		taskRunner := nats.NewAllTasksRunner(natsConn, localRunner, db, 10, viper.GetString("db-encryption-salt"))

		slog.Info("NATS worker started")
		return taskRunner.RunAll(cmd.Context())
	},
}

func init() {
	natsCmd.Flags().String("nats", "", "Nats connection string")
	if err := natsCmd.MarkFlagRequired("nats"); err != nil {
		panic(err)
	}

	natsCmd.Flags().String("database", "", "Database connection string")
	if err := natsCmd.MarkFlagRequired("database"); err != nil {
		panic(err)
	}

	natsCmd.Flags().String("db-encryption-salt", "", "Database salt encryption key")
	if err := natsCmd.MarkFlagRequired("db-encryption-salt"); err != nil {
		panic(err)
	}

	natsCmd.Flags().String("email-sendgrid", "", "Sendgrid API Key")

	natsCmd.Flags().String("email-sender", "no-reply@tedyst.ro", "Email sender")
	natsCmd.Flags().String("email-senderName", "Licenta", "Email sender name")

	rootCmd.AddCommand(natsCmd)
}
