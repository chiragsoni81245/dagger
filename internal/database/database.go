package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"

    "github.com/chiragsoni81245/dagger/internal/config"
)


func GetDB(config *config.Config) (*sql.DB, error) {
    var DB *sql.DB
	var err error

	// Build the connection string
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
		config.Database.User, config.Database.Password, config.Database.Host, config.Database.Port, config.Database.Name)

	// Connect to the database
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
        return nil, err
	}

	// Ping the database to check if the connection is successful
	if err = DB.Ping(); err != nil {
        return nil, err
	}

    return DB, err
}
