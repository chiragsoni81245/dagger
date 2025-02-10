package cmd

import (
	"fmt"
	"log"

	"github.com/chiragsoni81245/dagger/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback the last applied migration",
	Run: func(cmd *cobra.Command, args []string) {
        // Generate application configuration
        config, err := config.GetConfig(configPath)
        if err != nil {
            log.Fatal(err)
        }
        
        migrationsPath := fmt.Sprintf("file://%s", config.Server.MigrationsPath)
        databaseURI := fmt.Sprintf(
            "postgres://%s:%s@%s:%d/%s?sslmode=disable", 
            config.Database.User, 
            config.Database.Password, 
            config.Database.Host, 
            config.Database.Port, 
            config.Database.Name,
        )
		m, err := migrate.New(migrationsPath, databaseURI)
		if err != nil {
			log.Fatalf("Failed to initialize migration: %v", err)
		}

		err = m.Steps(-1) // Rollback the last migration
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Rollback failed: %v", err)
		}

		fmt.Println("Last migration rolled back successfully!")
	},
}

func init() {
	migrateCmd.AddCommand(downCmd)

	downCmd.Flags().StringVar(&configPath, "config", "", "Path to the config file (required)")
	downCmd.MarkFlagRequired("config")
}
