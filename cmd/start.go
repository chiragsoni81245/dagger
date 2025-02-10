package cmd

import (
	"fmt"
	"log"

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

var configPath string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the dagger service",
	Run: func(cmd *cobra.Command, args []string) {
        runServer(configPath)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVar(&configPath, "config", "", "Path to the config file (required)")
	startCmd.MarkFlagRequired("config")
}

