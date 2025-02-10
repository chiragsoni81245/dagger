package main

import (
	"fmt"
	"log"
    "os"

	"github.com/chiragsoni81245/dagger/internal/config"
	"github.com/chiragsoni81245/dagger/internal/server"
	"github.com/spf13/cobra"
)

func runServer(configPath string) {
    // Generate application configuration
    config, err := config.GetConfig(configPath)
    if err != nil {
        log.Fatal(err)
    }

    server, err := server.NewServer(config)
    if err != nil {
        log.Fatal(err)
    }

    server.Router.Run(fmt.Sprintf(":%d", config.Server.Port))
}

func main() {
    var configPath string

	// Root command
	var rootCmd = &cobra.Command{
		Use:   "dagger",
		Short: "Dagger is an orchestrator tool to run workflows",
		Run: func(cmd *cobra.Command, args []string) {
            runServer(configPath)
		},
	}

	// Define --config flag
	rootCmd.Flags().StringVar(&configPath, "config", "", "Path to the config file (required)")

	// Mark the flag as required
	rootCmd.MarkFlagRequired("config")

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
