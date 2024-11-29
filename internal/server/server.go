package server

import (
	"database/sql"
	"os"

	"github.com/chiragsoni81245/dagger/internal/config"
	"github.com/chiragsoni81245/dagger/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Server struct {
    Logger *logrus.Logger
    DB     *sql.DB
    Router *gin.Engine
}

func NewServer(config *config.Config) (*Server, error) {
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

    // Setup all routes
    SetupRoutes(r)


    return &Server{
        Logger: logger,
        DB: db,
        Router: r,
    }, nil
}
