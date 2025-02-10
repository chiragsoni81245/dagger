package cmd

import (
	"fmt"
	"log"

	"github.com/chiragsoni81245/dagger/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Required for PostgreSQL
	_ "github.com/golang-migrate/migrate/v4/source/file"       // Required for file-based migrations

	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply database migrations",
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

		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration failed: %v", err)
		}

		fmt.Println("Migrations applied successfully!")
	},
}

func init() {
	migrateCmd.AddCommand(upCmd)

	upCmd.Flags().StringVar(&configPath, "config", "", "Path to the config file (required)")
	upCmd.MarkFlagRequired("config")
}

