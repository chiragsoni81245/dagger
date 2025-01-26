package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/chiragsoni81245/dagger/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)


type WebSocketControllers struct {
    Server *types.Server
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin. Modify for stricter policies.
		return true
	},
}

func (wsC *WebSocketControllers) HandleWebSocket(c *gin.Context) {
	// Upgrade the HTTP connection to a WebSocket connection
    logger := wsC.Server.Logger

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Println("Failed to upgrade connection:", err)
		return
	}
	defer conn.Close()

    type webSocketMessage struct {
        Action string `json:"action"`
        EventDescriptor  string `json:"event"`
    }

	// Read messages from the client
	for {
        var message webSocketMessage;
		err := conn.ReadJSON(&message)
		if err != nil {
            if(websocket.IsCloseError(err, websocket.CloseGoingAway)){
                wsC.Server.Mux.Lock()
                delete(wsC.Server.EventSubscriptions, conn)
                wsC.Server.Mux.Unlock()
                break
            }
			logger.Errorln("Error reading message:", err)
			break
		}
        
        eventParts := strings.Split(message.EventDescriptor, ":")
        if len(eventParts) != 2 {
            logger.Errorf("Invalid event")
            break 
        }
        resourceName := eventParts[0]
        resourceId, err := strconv.Atoi(eventParts[1])
        if err != nil {
            logger.Errorln("Error in websocket message event parsing:", err)
            break
        }

        wsC.Server.Mux.Lock()
        if message.Action == "subscribe" {
            if _, ok := wsC.Server.EventSubscriptions[conn]; ok {
                wsC.Server.EventSubscriptions[conn] = append(wsC.Server.EventSubscriptions[conn], types.EventDescriptor{
                   Resource: resourceName,
                   ID: int(resourceId),
                })
            } else {
                wsC.Server.EventSubscriptions[conn] = []types.EventDescriptor{
                    {
                        Resource: resourceName,
                        ID: int(resourceId),
                    },
                }
            }
        } else if message.Action == "unsubscribe" {
            if _, ok := wsC.Server.EventSubscriptions[conn]; ok {
                newEvents := []types.EventDescriptor{}
                for _, event := range wsC.Server.EventSubscriptions[conn] {
                    if fmt.Sprintf("%s:%d", event.Resource, event.ID) != message.EventDescriptor {
                        newEvents = append(newEvents, types.EventDescriptor{Resource: resourceName, ID: resourceId})
                    }
                }
                wsC.Server.EventSubscriptions[conn] = newEvents
            }
        }
        wsC.Server.Mux.Unlock()
	}
}
