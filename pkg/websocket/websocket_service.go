//  There is a single Manager instance that manages all WebSocket connections, and each connected user is represented as a Client.

// The Manager maintains a list of active clients (ClientList), routes events, and ensures clients are added/removed correctly.
// Each Client has its own WebSocket connection, message channels, and is associated with a room for room-wise communication.
// Clients send messages to their room, and only users in the same room receive those messages.
// This structure ensures that room-based messaging works efficiently while maintaining a centralized manager for all WebSocket connections.

package websocket

import (
	logs "brainwars/pkg/logger"
	"brainwars/pkg/quiz"
	quizmodel "brainwars/pkg/quiz/model"
	"brainwars/pkg/room"
	roommodel "brainwars/pkg/room/model"
	user "brainwars/pkg/users"
	usermodel "brainwars/pkg/users/model"
	"brainwars/pkg/util"
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
	gameStates map[string]*quizmodel.GameState
}

type ClientList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	manager    *Manager
	egress     chan Event
	botEvents  chan Event // Only used for bot clients
	roomCode   string
	isBot      bool              // Flag to identify bot clients
	botType    usermodel.BotType // Empty for real users, "30sec", "1min", "2min" for bots
	userID     uuid.UUID         // Store the user ID for easier reference
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
		gameStates: make(map[string]*quizmodel.GameState),
	}
	m.setupEventHandlers()
	return m
}

func NewClient(conn *websocket.Conn, manager *Manager, roomCode string, isBot bool, botType usermodel.BotType, userID uuid.UUID) *Client {
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

}

// when user joins a room this serveWs handler is called
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
	roomCode := c.Query("roomCode")
	if roomCode == "" {
		roomCode = "8bd9c332-ea09-434c-b439-5b3a39d3de5f" // Default room for testing
	}

	userID := util.GetUserIDFromctx(ctx)
	roomMember, err := room.GetRoomMemberByRoomAndUserID(ctx, roommodel.RoomMemberReq{
		UserID: userID,
		RoomID: uuid.MustParse(roomCode),
	})
	if err != nil {
		l.Sugar().Error("get room member by room and user id failed", err)
		return
	}

	totalQuestions := 10 // TODO: need to come from db
	if roomMember == nil {
		l.Sugar().Error("room member is nil")
		return
	}

	client := NewClient(conn, m, roomCode, false, "", userID)
	m.addClient(client)

	// Check if the room needs to be initialized
	m.initializeRoomGameState(ctx, roomCode, totalQuestions)

	// When a human player joins, set bots to ready state
	// even if 1 user joins the room then we instentaniously set up all the bots to ready state for the game to start.
	// we get the list of bots from list all members in a room where bots are members as ready
	err = m.setupUserForRoom(ctx, roomCode, userID)
	if err != nil {
		return
	}
	m.setupBotsForRoom(ctx, roomCode)

	go client.readMessages(ctx)
	go client.writeMessages(ctx)
}

// Set up bots to be ready when a human player joins
func (m *Manager) setupUserForRoom(ctx context.Context, roomCode string, userID uuid.UUID) error {
	l := logs.GetLoggerctx(ctx)

	userDetails, err := user.GetUserDetailsbyID(ctx, userID)
	if err != nil {
		l.Sugar().Error("get user details by id failed", err)
		return err
	}

	// Set the user to ready state

	// Notify all clients that this user is ready
	botReadyNotification := Payload{
		Data: fmt.Sprintf("User %s is ready", userDetails.UserName),
		Time: time.Now(),
	}

	data, err := json.Marshal(botReadyNotification)
	if err != nil {
		l.Sugar().Error("bot ready notification json marshal failed", err)
		return err
	}

	readyEvent := Event{Type: "game_status", Payload: data}

	// Broadcast to all clients in the room
	for client := range m.clients[roomCode] {
		client.egress <- readyEvent
	}

	return nil
}

// Start game handler
func StartGameMessageHandler(ctx context.Context, event Event, c *Client) error {
	l := logs.GetLoggerctx(ctx)

	// Get the game state for this room
	c.manager.Lock()
	gameState, exists := c.manager.gameStates[c.roomCode]
	if !exists {
		c.manager.Unlock()
		return fmt.Errorf("game state not found for room %s", c.roomCode)
	}

	// Update game state
	gameState.RoomStatus = roommodel.Started
	gameState.StartTime = time.Now()
	gameState.CurrentQuestionIndex = 0
	c.manager.Unlock()

	// Fetch questions from the database
	questions, err := quiz.ListQuestionsByRoomCode(ctx, c.roomCode)
	if err != nil {
		l.Sugar().Error("failed to fetch questions:", err)
		return err
	}

	// Store questions in game state
	c.manager.Lock()
	gameState.Questions = questions // TODO: remove it
	c.manager.Unlock()

	// Send the first question to all clients
	return sendNextQuestion(ctx, c.manager, c.roomCode)
}

// Send the next question to all clients in a room
func sendNextQuestion(ctx context.Context, manager *Manager, roomCode string) error {
	l := logs.GetLoggerctx(ctx)

	manager.Lock()
	gameState, exists := manager.gameStates[roomCode]
	if !exists {
		manager.Unlock()
		l.Sugar().Error("game state not found for room %s", roomCode)
		return fmt.Errorf("game state not found for room %s", roomCode)
	}

	// Check if we've reached the end of questions
	if gameState.CurrentQuestionIndex >= len(gameState.Questions) {
		// Game is over
		gameState.RoomStatus = roommodel.Ended
		manager.Unlock()

		// Send game end event
		// leaderboard:
		endGamePayload := struct {
			Message    string                  `json:"message"`
			Scores     []quizmodel.Participant `json:"scores"`
			FinishTime time.Time               `json:"finishTime"`
		}{
			Message:    "Game has ended. Here are the final scores.",
			Scores:     gameState.Participants,
			FinishTime: time.Now(),
		}

		endGameData, _ := json.Marshal(endGamePayload)
		endEvent := Event{Type: EventEndGame, Payload: endGameData}

		// Broadcast to all clients including bots
		for client := range manager.clients[roomCode] {
			client.egress <- endEvent
		}

		// Also notify bots about game end
		manager.broadcastToBots(ctx, roomCode, endEvent)

		return nil
	}

	// Get the current question
	currentQuestion := gameState.Questions[gameState.CurrentQuestionIndex]

	// Create a client-safe version (without correct answer)
	clientQuestion := Question{
		ID:        currentQuestion.ID,
		Question:  currentQuestion.Question,
		Options:   currentQuestion.Options,
		TimeLimit: currentQuestion.TimeLimit,
	}

	manager.Unlock()

	// Prepare question event
	questionData, _ := json.Marshal(struct {
		QuestionIndex  int       `json:"questionIndex"`
		TotalQuestions int       `json:"totalQuestions"`
		Question       Question  `json:"question"`
		StartTime      time.Time `json:"startTime"`
	}{
		QuestionIndex:  gameState.CurrentQuestionIndex + 1,
		TotalQuestions: len(gameState.Questions),
		Question:       clientQuestion,
		StartTime:      time.Now(),
	})

	questionEvent := Event{Type: EventNewQuestion, Payload: questionData}

	// Broadcast to all clients
	for client := range manager.clients[roomCode] {
		client.egress <- questionEvent
	}

	// Notify bots about new question so they can prepare to answer
	manager.broadcastToBots(ctx, roomCode, questionEvent)

	// Schedule next question after a delay (current question time limit + 5 seconds for results)
	go func() {
		timeLimit := time.Duration(currentQuestion.TimeLimit+5) * time.Second
		timer := time.NewTimer(timeLimit)
		<-timer.C

		// Short delay to let players see results
		time.Sleep(3 * time.Second)

		// Move to next question
		manager.Lock()
		if gameState, exists := manager.gameStates[roomCode]; exists {
			gameState.CurrentQuestionIndex++
		}
		manager.Unlock()

		// Send next question
		sendNextQuestion(ctx, manager, roomCode)
	}()

	return nil
}

// Handle answer submissions
func SubmitAnswerHandler(ctx context.Context, event Event, c *Client) error {
	l := logs.GetLoggerctx(ctx)

	var submission AnswerSubmission
	if err := json.Unmarshal(event.Payload, &submission); err != nil {
		return fmt.Errorf("bad payload: %v", err)
	}

	// Set timestamp if not provided
	if submission.Timestamp.IsZero() {
		submission.Timestamp = time.Now()
	}

	// Get the game state
	c.manager.Lock()
	gameState, exists := c.manager.gameStates[c.roomCode]
	if !exists {
		c.manager.Unlock()
		return fmt.Errorf("game state not found for room %s", c.roomCode)
	}

	// Only process if game is in progress
	if gameState.RoomStatus != "in_progress" || gameState.CurrentQuestionIndex >= len(gameState.Questions) {
		c.manager.Unlock()
		return fmt.Errorf("game is not in active question phase")
	}

	// Get current question
	currentQuestion := gameState.Questions[gameState.CurrentQuestionIndex]

	// Check if answer is correct
	isCorrect := submission.Answer == currentQuestion.CorrectAnswer

	// Update participant score
	found := false
	for i, participant := range gameState.Participants {
		if participant.UserID == c.userID {
			if isCorrect {
				// Calculate score based on answer speed
				answerTime := submission.Timestamp.Sub(gameState.StartTime)
				speedBonus := float64(currentQuestion.TimeLimit) - answerTime.Seconds()
				if speedBonus < 0 {
					speedBonus = 0
				}

				// Add score (base 100 + speed bonus up to 100)
				gameState.Participants[i].Score += 100 + int(speedBonus*3.33)
			}
			found = true
			break
		}
	}

	// If participant not found, add them
	if !found {
		// Get user info from database
		var username string
		// TODO: Replace with actual database query to get username
		if c.isBot {
			username = fmt.Sprintf("Bot-%s", c.userID.String()[:8])
		} else {
			username = fmt.Sprintf("User-%s", c.userID.String()[:8])
		}

		score := 0
		if isCorrect {
			score = 100 // Base score for correct answer
		}

		gameState.Participants = append(gameState.Participants, Participant{
			UserID:   c.userID,
			Username: username,
			IsBot:    c.isBot,
			Score:    score,
			IsReady:  true,
		})
	}

	c.manager.Unlock()

	// Acknowledge answer submission
	ackPayload := struct {
		QuestionID uuid.UUID `json:"questionId"`
		Received   bool      `json:"received"`
		IsCorrect  bool      `json:"isCorrect"`
	}{
		QuestionID: currentQuestion.ID,
		Received:   true,
		IsCorrect:  isCorrect,
	}

	ackData, _ := json.Marshal(ackPayload)
	ackEvent := Event{Type: "answer_received", Payload: ackData}

	// Send acknowledgment only to the client who submitted
	c.egress <- ackEvent

	return nil
}

// Modified ReadyGameMessageHandler to check if all participants are ready and start game
func ReadyGameMessageHandler(ctx context.Context, event Event, c *Client) error {
	l := logs.GetLoggerctx(ctx)

	userDetails, err := user.GetUserDetailsbyID(ctx, c.userID)
	if err != nil {
		l.Sugar().Error("get user details by id failed", err)
		return err
	}
	if userDetails == nil {
		return fmt.Errorf("user not found")
	}
	// Update the room member's ready status
	err = room.UpdateRoomMemberByID(ctx, roommodel.RoomMemberReq{
		UserID:           c.userID,
		RoomID:           uuid.MustParse(c.roomCode),
		RoomMemberStatus: roommodel.ReadyQuiz,
	})
	if err != nil {
		l.Sugar().Error("Failed to update ready status", err)
		return err
	}

	// Check if all room members are ready
	roomMembers, err := room.ListRoomMembersByRoomID(ctx, roommodel.RoomIDReq{
		RoomID: uuid.MustParse(c.roomCode),
	})
	if err != nil {
		l.Sugar().Error("List Room member by room id failed", err)
		return err
	}

	// Check if everyone is ready
	allReady := true
	for _, member := range roomMembers {
		if member.RoomMemberStatus != roommodel.ReadyQuiz {
			allReady = false
			break
		}
	}

	if allReady {
		// Everyone is ready, start the game
		gameReadyNotification := struct {
			Status  string    `json:"status"`
			Message string    `json:"message"`
			StartAt time.Time `json:"startAt"`
		}{
			Status:  "start_game",
			Message: "All players are ready. Game begins.",
			// TODO: Cross check this viper
			StartAt: time.Now().Add(time.Duration(viper.GetInt("game.gamestartbuffer")) * time.Second), // Start after 3 seconds
		}

		startData, _ := json.Marshal(gameReadyNotification)
		startEvent := Event{Type: EventStartGame, Payload: startData}

		// Broadcast start notification to all clients
		for client := range c.manager.clients[c.roomCode] {
			client.egress <- startEvent
		}

		// Wait 3 seconds then start the game
		go func() {
			time.Sleep(time.Duration(viper.GetInt("game.gamestartbuffer")) * time.Second)
			StartGameMessageHandler(ctx, startEvent, c)
		}()
	}

	return nil
}

// Initialize the game state for a room
func (m *Manager) initializeRoomGameState(ctx context.Context, roomCode string, totalRounds int) {
	m.Lock()
	defer m.Unlock()

	if _, exists := m.gameStates[roomCode]; !exists {
		// Initialize a new game state for this room
		m.gameStates[roomCode] = &quizmodel.GameState{
			RoomCode:     roomCode,
			RoomStatus:   roommodel.Waiting, // waiting for players
			CurrentRound: 0,
			TotalRounds:  totalRounds, // Default to 10 rounds
			Questions:    []quizmodel.Question{},
			Participants: []quizmodel.Participant{},
		}
	}
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

// payload structure {data:ready_game, time:time.now()}

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
