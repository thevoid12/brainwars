package websocket

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	/**
	websocketUpgrader is used to upgrade incomming HTTP requests into a persitent websocket connection
	*/
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

// Manager is used to hold references to all Clients Registered, and Broadcasting etc
type Manager struct {
}

// NewManager is used to initalize all the values inside the manager
func NewManager() *Manager {
	return &Manager{}
}

// ServeWS is a HTTP Handler that the has the Manager that allows connections
func (m *Manager) ServeWS(c *gin.Context) {
	log.Println("New WebSocket connection")

	// Upgrade the connection
	conn, err := websocketUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	defer conn.Close()
}
