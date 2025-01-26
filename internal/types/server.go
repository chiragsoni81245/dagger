package types

import (
	"database/sql"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type Server struct {
    Logger *logrus.Logger
    DB     *sql.DB
    Router *gin.Engine
    EventSubscriptions map[*websocket.Conn][]EventDescriptor
    Mux sync.Mutex
    EventCh chan Event
}
