package main

import (
	"fmt"
	"log"

	"github.com/chiragsoni81245/dagger/internal/config"
	"github.com/chiragsoni81245/dagger/internal/server"
)

func main() {
    // Generate application configuration
    config, err := config.GetConfig()
    if err != nil {
        log.Fatal(err)
    }

    server, err := server.NewServer(config)
    if err != nil {
        log.Fatal(err)
    }

    server.Router.Run(fmt.Sprintf(":%d", config.Server.Port))
}
