//  There is a single Manager instance that manages all WebSocket connections, and each connected user is represented as a Client.

// The Manager maintains a list of active clients (ClientList), routes events, and ensures clients are added/removed correctly.
// Each Client has its own WebSocket connection, message channels, and is associated with a room for room-wise communication.
// Clients send messages to their room, and only users in the same room receive those messages.
// This structure ensures that room-based messaging works efficiently while maintaining a centralized manager for all WebSocket connections.

package websocket

import (
	logs "brainwars/pkg/logger"
	"brainwars/pkg/room"
	roommodel "brainwars/pkg/room/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
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
	handlers   map[string]EventHandler
	roomStates map[string]*roommodel.RoomStatus
}

type ClientList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	manager    *Manager
	egress     chan Event
	botEvents  chan Event // Only used for bot clients
	roomCode   string
	isBot      bool      // Flag to identify bot clients
	botType    string    // Empty for real users, "30sec", "1min", "2min" for bots
	userID     uuid.UUID // Store the user ID for easier reference
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
		clients:    make(map[string]ClientList),
		handlers:   make(map[string]EventHandler),
		roomStates: make(map[string]*roommodel.RoomStatus),
	}
	m.setupEventHandlers()
	return m
}

func NewClient(conn *websocket.Conn, manager *Manager, roomCode string, isBot bool, botType string, userID uuid.UUID) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan Event),
		botEvents:  make(chan Event),
		roomCode:   roomCode,
		isBot:      isBot,
		botType:    botType,
		userID:     userID,
	}
	i
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
	isBot := false
	botType := ""
	userID := uuid.New()
	client := NewClient(conn, m, roomCode, isBot, botType, userID)
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

	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		l.Sugar().Error("set read deadline failed", err)
		return
	}
	// Configure how to handle Pong responses
	c.connection.SetPongHandler(c.pongHandler)
	c.connection.SetReadLimit(viper.GetInt64("ws.maxReadSize")) // max read size limit bytes
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
		if err := c.manager.routeEvent(ctx, request, c); err != nil {
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
			log.Println("ping")
			c.connection.WriteMessage(websocket.PingMessage, nil)
		}
	}
}

// pongHandler is used to handle PongMessages for the Client
func (c *Client) pongHandler(pongMsg string) error {
	// Current time + Pong Wait time
	log.Println("pong")
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}

func (m *Manager) routeEvent(ctx context.Context, event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		return handler(ctx, event, c)
	}
	return ErrEventNotSupported
}

// functions which we use after creating websocket manager
func SendMessageHandler(ctx context.Context, event Event, c *Client) error {
	var msgEvent Payload
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

func ReadyGameMessageHandler(ctx context.Context, event Event, c *Client) error {
	l := logs.GetLoggerctx(ctx)

	var msgEvent Payload
	if err := json.Unmarshal(event.Payload, &msgEvent); err != nil {
		return fmt.Errorf("bad payload: %v", err)
	}
	err := room.UpdateRoomMemberByID(ctx, roommodel.RoomMemberReq{
		ID:               uuid.UUID{},
		UserID:           uuid.UUID{},
		RoomID:           uuid.UUID{},
		RoomMemberStatus: roommodel.ReadyQuiz,
		RoomCode:         "",
	})
	if err != nil {
		l.Sugar().Error("update room member by id failed", err)
		return err
	}

	var gameStatus Payload
	gameStatus = Payload{
		Data: "the user +____+ is ready", // TODO: fill in the user
		Time: time.Now(),
	}
	data, err := json.Marshal(gameStatus)
	if err != nil {
		l.Sugar().Error("marshal game status failed", err)
		return err
	}

	// letting everyone know that this client is ready
	outgoing := Event{Type: "game_status", Payload: data}

	for client := range c.manager.clients[c.roomCode] {
		client.egress <- outgoing
	}

	// get room member info's if all are ready then start the game
	roomMembers, err := room.ListRoomMembersByRoomID(ctx, roommodel.RoomIDReq{
		UserID: uuid.UUID{},
		RoomID: uuid.UUID{},
	})
	if err != nil {
		l.Sugar().Error("List Room member by room id failed", err)
		return err
	}
	// if all okay start the game and get the first question and
	// broadcast to all the users including all the bots as well
	isokay := true
	for _, roommember := range roomMembers {
		if (!roommember.IsBot) && roommember.RoomMemberStatus != roommodel.ReadyQuiz {
			isokay = false
			break
		}
	}

	if isokay { // everything  is cool everyone is ready so we start the game
		gameReadyNotification := struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}{
			Status:  "start_game",
			Message: "All players are ready. Game can begin.",
		}

		notifyData, err := json.Marshal(gameReadyNotification)
		if err != nil {
			l.Sugar().Errorf("Failed to marshal game ready notification: %v", err)
			return nil
		}

		// Broadcast ready-to-start notification
		readyEvent := Event{Type: "start_game", Payload: notifyData}
		for client := range c.manager.clients[c.roomCode] {
			client.egress <- readyEvent
		}
	}
	return nil
}

// func ChatRoomHandler(event Event, c *Client) error {
// 	var changeRoomEvent ChangeRoomEvent
// 	if err := json.Unmarshal(event.Payload, &changeRoomEvent); err != nil {
// 		return fmt.Errorf("bad payload: %v", err)
// 	}
// 	c.manager.Lock()
// 	c.manager.removeClient(c)
// 	c.roomCode = changeRoomEvent.Name
// 	c.manager.addClient(c)
// 	c.manager.Unlock()
// 	return nil
// }
