package server

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/chiragsoni81245/dagger/internal/config"
	"github.com/chiragsoni81245/dagger/internal/database"
	"github.com/chiragsoni81245/dagger/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

func eventOrchestration(server *types.Server) {
    for {
        e := <-server.EventCh
        eJson, err := json.Marshal(e)
        if err != nil {
            server.Logger.Errorf("Error parsing event: %v", err)
        }
        event := make(map[string]any)
        err = json.Unmarshal(eJson, &event)
        if err != nil {
            server.Logger.Errorf("Error parsing event: %v", err)
        }

        server.Mux.Lock()
        for conn, eventDescriptors := range server.EventSubscriptions {
            for _, eventDescriptor := range eventDescriptors {
                if (eventDescriptor.Resource == e.Resource &&
                eventDescriptor.ID == e.ID) || (e.ParentResource != nil &&
                eventDescriptor.Resource == e.ParentResource.Resource &&
                eventDescriptor.ID == e.ParentResource.ID) {
                    event["eventDescriptor"] = fmt.Sprintf("%s:%d", eventDescriptor.Resource, eventDescriptor.ID)
                    err := conn.WriteJSON(event)
                    if err != nil {
                        server.Logger.Errorf("Error sending event: %v", err)
                    }
                }
            }
        }
        server.Mux.Unlock()
    }
}

func NewServer(config *config.Config) (*types.Server, error) {
	// Create a global logger further will be used in whole application
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logLevel, err := logrus.ParseLevel(config.Server.LogLevel)
	if err != nil {
		return nil, err
	}
	logger.SetLevel(logLevel)

	// Create a global db connection which will be used in whole application
	db, err := database.GetDB(config)
	if err != nil {
		return nil, err
	}

	r := gin.Default()

	r.Static("/static", "internal/server/public")

	// Use middleware to inject logger and DB
	r.Use(InjectDependencies(logger, db))

	eventSubscriptions := make(map[*websocket.Conn][]types.EventDescriptor)
	eventCh := make(chan types.Event)

	server := types.Server{
		Logger:             logger,
		DB:                 db,
		Router:             r,
		EventSubscriptions: eventSubscriptions,
		EventCh:            eventCh,
	}

	go eventOrchestration(&server)

	// Setup all routes
	SetupRoutes(&server)

	return &server, nil
}
