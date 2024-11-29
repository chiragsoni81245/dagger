package server

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func InjectDependencies(logger *logrus.Logger, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add logger to context
		c.Set("logger", logger)
		// Add database connection to context
		c.Set("db", db)
		c.Next() // Pass to the next middleware/handler
	}
}
