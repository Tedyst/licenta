package cmd

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tedyst/licenta/bruteforce"
	database "github.com/tedyst/licenta/db"
	"github.com/tedyst/licenta/email"
	localExchange "github.com/tedyst/licenta/messages/local"
	"github.com/tedyst/licenta/scheduler"
	"github.com/tedyst/licenta/tasks/local"
)

var schedulerCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "Automatic scheduler",
	Long:  `Schedules scans and updates for the database`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db := database.InitDatabase(viper.GetString("database"))

		hashKey, err := hex.DecodeString(viper.GetString("hash-key"))
		if err != nil {
			return fmt.Errorf("hash key must be a hex string: %w", err)
		}
		if len(hashKey) != 32 {
			return fmt.Errorf("hash key must be 32 bytes long (64 hex characters)")
		}

		encryptKey, err := hex.DecodeString(viper.GetString("encrypt-key"))
		if err != nil {
			return fmt.Errorf("encrypt key must be a hex string: %w", err)
		}
		if len(encryptKey) != 32 {
			return fmt.Errorf("encrypt key must be 32 bytes long (64 hex characters)")
		}

		localExchange := localExchange.NewLocalExchange()
		bruteforceProvider := bruteforce.NewDatabaseBruteforceProvider(db)

		emailSender := email.NewSendGridEmailSender(
			viper.GetString("email-sendgrid"),
			viper.GetString("email-senderName"),
			viper.GetString("email-sender"),
		)

		taskRunner := local.NewLocalRunner(
			viper.GetBool("debug"),
			emailSender,
			db,
			localExchange,
			bruteforceProvider,
			viper.GetString("db-encryption-salt"),
		)

		sc := scheduler.NewScheduler(db, taskRunner)
		err = sc.RunContinuous(cmd.Context(), 24*time.Hour)
		if err != nil {
			return fmt.Errorf("could not run scheduler: %w", err)
		}

		return nil
	},
}

func init() {
	schedulerCmd.Flags().String("database", "", "Database connection string")

	schedulerCmd.Flags().String("hash-key", "", "Hash key used for signing Cookies")
	if err := schedulerCmd.MarkFlagRequired("hash-key"); err != nil {
		panic(err)
	}
	schedulerCmd.Flags().String("encrypt-key", "", "Encrypt key used for signing Cookies")
	if err := schedulerCmd.MarkFlagRequired("encrypt-key"); err != nil {
		panic(err)
	}
	schedulerCmd.Flags().String("db-encryption-salt", "", "Database salt encryption key")
	if err := schedulerCmd.MarkFlagRequired("db-encryption-salt"); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(schedulerCmd)
}
