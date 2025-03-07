package handlers

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins (configure this properly in prod)
	},
}

var rooms sync.Map // RoomCode -> map[*websocket.Conn]bool

func handleWebSocket(c *gin.Context) {
	roomCode := c.Query("roomCode")
	if roomCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "roomCode required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	clients, _ := rooms.LoadOrStore(roomCode, &sync.Map{})
	clientMap := clients.(*sync.Map)
	clientMap.Store(conn, true)
	fmt.Println("Client connected to room:", roomCode)

	go pingPong(conn, roomCode)
	if c.Query("bot") == "true" {
		go simulateBot(conn, roomCode)
	}

	readMessages(conn, roomCode)
}

func readMessages(conn *websocket.Conn, roomCode string) {
	clients, _ := rooms.Load(roomCode)
	clientMap := clients.(*sync.Map)

	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected from room:", roomCode)
			clientMap.Delete(conn)
			break
		}
		broadcast(roomCode, messageType, msg)
	}
}

func broadcast(roomCode string, messageType int, msg []byte) {
	clients, _ := rooms.Load(roomCode)
	clientMap := clients.(*sync.Map)

	clientMap.Range(func(key, value interface{}) bool {
		client := key.(*websocket.Conn)
		if err := client.WriteMessage(messageType, msg); err != nil {
			fmt.Println("Write error:", err)
			client.Close()
			clientMap.Delete(client)
		}
		return true
	})
}

func pingPong(conn *websocket.Conn, roomCode string) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
			fmt.Println("Ping failed, closing connection:", err)
			conn.Close()
			break
		}
	}
}

func simulateBot(conn *websocket.Conn, roomCode string) {
	time.Sleep(2 * time.Second)
	msg := []byte("Bot Answer")
	conn.WriteMessage(websocket.TextMessage, msg)
}
