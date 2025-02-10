package cmd

import "github.com/spf13/cobra"

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Manage database migrations",
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

