package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dagger",
	Short: "Dagger CLI an orchestration tool to run manage workflows",
	Long:  `Dagger CLI an orchestration tool to run manage workflows`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
