package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "doghole",
	Short: "Doghole is a command-line tool for managing and interacting with the Doghole system.",
	Long: `Doghole is a command-line tool designed to facilitate the management and interaction with the Doghole system.
It provides various commands to perform operations such as starting the server, managing configurations, and handling database interactions.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help() // Display help if no subcommand is provided
	},
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
