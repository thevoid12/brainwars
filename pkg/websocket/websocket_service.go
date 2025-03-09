//  There is a single Manager instance that manages all WebSocket connections, and each connected user is represented as a Client.

// The Manager maintains a list of active clients (ClientList), routes events, and ensures clients are added/removed correctly.
// Each Client has its own WebSocket connection, message channels, and is associated with a room for room-wise communication.
// Clients send messages to their room, and only users in the same room receive those messages.
// This structure ensures that room-based messaging works efficiently while maintaining a centralized manager for all WebSocket connections.

package websocket

import (
	logs "brainwars/pkg/logger"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var websocketUpgrader = websocket.Upgrader{
	// 	CheckOrigin:     checkOrigin, TODO: SETUP ORIGIN CHECK
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var ErrEventNotSupported = errors.New("this event type is not supported")

func checkOrigin(r *http.Request) bool {
	return r.Header.Get("Origin") == "https://localhost:8080" // TODO: move it to viper config
}

type Manager struct {
	clients map[string]ClientList // map key is roomCode value is the list of clients in that room
	sync.RWMutex
	handlers map[string]EventHandler
}

type ClientList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	manager    *Manager
	egress     chan Event
	roomCode   string
}

type NewMessageEvent struct {
	Payload
	Sent time.Time `json:"sent"`
}

type ChangeRoomEvent struct {
	Name string `json:"name"`
}

func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		clients:  make(map[string]ClientList),
		handlers: make(map[string]EventHandler),
	}
	m.setupEventHandlers()
	return m
}

func NewClient(conn *websocket.Conn, manager *Manager, roomCode string) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan Event),
		roomCode:   roomCode,
	}
}

func (m *Manager) ServeWS(c *gin.Context) {
	ctx := c.Request.Context()
	l := logs.GetLoggerctx(ctx)

	log.Println("New connection")
	conn, err := websocketUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		l.Sugar().Error("websocket upgrade error:", err)
		return
	}
	// get the roomID from the query params
	// roomID := c.Query("roomID")
	roomCode := "8bd9c332-ea09-434c-b439-5b3a39d3de5f"
	client := NewClient(conn, m, roomCode)
	m.addClient(client)
	go client.readMessages(ctx)
	go client.writeMessages(ctx)
}

// mapping clients to the roomID
func (m *Manager) addClient(client *Client) {
	m.Lock()
	if _, exists := m.clients[client.roomCode]; !exists {
		m.clients[client.roomCode] = make(ClientList)
	}
	m.clients[client.roomCode][client] = true
	m.Unlock()
}

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	if _, exists := m.clients[client.roomCode]; exists {
		delete(m.clients[client.roomCode], client)
		client.connection.Close()
	}
	m.Unlock()
}

func (c *Client) readMessages(ctx context.Context) {
	l := logs.GetLoggerctx(ctx)
	defer c.manager.removeClient(c)
	for {
		_, payload, err := c.connection.ReadMessage()
		if err != nil {
			l.Sugar().Error("error reading message:", err)
			break
		}
		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			l.Sugar().Error("error unmarshalling message:", err)
			break
		}
		if err := c.manager.routeEvent(request, c); err != nil {
			l.Sugar().Error("error routing event:", err)
			break
		}
	}
}

func (c *Client) writeMessages(ctx context.Context) {
	l := logs.GetLoggerctx(ctx)

	l.Info("Client connected for write messages")
	ticker := time.NewTicker(time.Second * 9)
	defer func() {
		ticker.Stop()
		c.manager.removeClient(c)
	}()
	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				c.connection.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			data, _ := json.Marshal(message)
			c.connection.WriteMessage(websocket.TextMessage, data)
		case <-ticker.C:
			c.connection.WriteMessage(websocket.PingMessage, nil)
		}
	}
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		return handler(event, c)
	}
	return ErrEventNotSupported
}

// functions which we use after creating websocket manager
func SendMessageHandler(event Event, c *Client) error {
	var msgEvent Payload
	fmt.Println(event.Payload)
	if err := json.Unmarshal(event.Payload, &msgEvent); err != nil {
		return fmt.Errorf("bad payload: %v", err)
	}

	// here we do all our logic and fill in the unbuffered channel
	broadMessage := NewMessageEvent{Payload: msgEvent, Sent: time.Now()}
	data, _ := json.Marshal(broadMessage)
	outgoing := Event{Type: "send_message", Payload: data}

	for client := range c.manager.clients[c.roomCode] {
		client.egress <- outgoing
	}
	return nil
}

func ChatRoomHandler(event Event, c *Client) error {
	var changeRoomEvent ChangeRoomEvent
	if err := json.Unmarshal(event.Payload, &changeRoomEvent); err != nil {
		return fmt.Errorf("bad payload: %v", err)
	}
	c.manager.Lock()
	c.manager.removeClient(c)
	c.roomCode = changeRoomEvent.Name
	c.manager.addClient(c)
	c.manager.Unlock()
	return nil
}
