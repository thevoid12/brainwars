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
	"brainwars/pkg/room/model"
	roommodel "brainwars/pkg/room/model"
	user "brainwars/pkg/users"
	usermodel "brainwars/pkg/users/model"
	"brainwars/pkg/util"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
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
	room       *roommodel.Room
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

func NewClient(conn *websocket.Conn, manager *Manager, roomCode string, isBot bool, botType usermodel.BotType, userID uuid.UUID, room *model.Room) *Client {

	// // Only set up pong handler for real clients with WebSocket connections
	// if conn != nil {
	// 	conn.SetPongHandler(client.pongHandler)
	// }
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan Event),
		botEvents:  make(chan Event),
		roomCode:   roomCode,
		isBot:      isBot,
		botType:    botType,
		userID:     userID,
		room:       room,
	}

}

// when user joins a room this serveWs handler is called
// TODO: render error templates
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
		l.Sugar().Error("room code not found", err)
		return
	}

	userInfo := util.GetUserInfoFromctx(ctx)
	userID := userInfo.ID

	roomDetails, err := room.GetRoomByRoomCode(ctx, roomCode)
	if err != nil {
		l.Sugar().Error("get room by room code failed", err)
		return
	}
	if roomDetails == nil {
		l.Sugar().Error("room not found")
		return
	}

	roomMember, err := room.GetRoomMemberByRoomCodeAndUserID(ctx, roommodel.RoomMemberReq{
		UserID:   userID,
		RoomCode: roomCode,
	})
	if err != nil {
		l.Sugar().Error("get room member by room and user id failed", err)
		return
	}
	if roomMember == nil {
		l.Sugar().Error("there are no room mebers")
		return
	}

	questions, err := quiz.ListQuestionsByRoomCode(ctx, roomCode)
	if err != nil {
		l.Sugar().Error("list questions by room code failed", err)
		return
	}

	totalQuestions := questions.QuestionCount

	client := NewClient(conn, m, roomCode, false, "", userID, roomDetails)
	m.addClient(client)

	// Check if the room needs to be initialized
	m.initializeRoomGameState(ctx, roomCode, totalQuestions)
	go client.readMessages(ctx)
	go client.writeMessages(ctx)
	// When a human player joins, set bots to ready state
	// even if 1 user joins the room then we instentaniously set up all the bots to ready state for the game to start.
	// we get the list of bots from list all members in a room where bots are members as ready
	err = m.setupUserForRoom(ctx, roomCode, userID)
	if err != nil {
		return
	}

	m.setupBotsForRoom(ctx, roomCode, roomDetails)
	// if the game is a single player game since the user is ready and bots are ready as well
	//  we automatically display the first question. in terms of multiplayer game a button needs to be triggered
	// to start the game
	if roomDetails.GameType == roommodel.SP {
		err = StartGameMessageHandler(ctx, Event{}, client)
		if err != nil {
			return
		}
	}

	return
}

// Set up bots to be ready when a human player joins
func (m *Manager) setupUserForRoom(ctx context.Context, roomCode string, userID uuid.UUID) error {
	l := logs.GetLoggerctx(ctx)
	// Use context with timeout for database operation
	// dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancel()

	var userDetails *usermodel.UserInfo
	var err error

	// Retry logic for getting user details
	for retries := 0; retries < 3; retries++ {
		userDetails, err = user.GetUserDetailsbyID(ctx, userID)
		if err == nil {
			break
		}
		l.Sugar().Warnf("Get user details attempt %d failed: %v. Retrying...", retries+1, err)
		time.Sleep(1 * time.Second)
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

	readyEvent := Event{Type: "ready_game", Payload: data}

	// TODO: update user room member's state to ready
	// for retries := 0; retries < 3; retries++ {
	// 	userDetails, err = room.UpdateRoomMemberByID(ctx)
	// 	if err == nil {
	// 		break
	// 	}
	// 	l.Sugar().Warnf("Get user details attempt %d failed: %v. Retrying...", retries+1, err)
	// 	time.Sleep(1 * time.Second)
	// }

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
	if gameState.CurrentQuestionIndex >= len(gameState.Questions.QuestionData) {
		// Game is over
		gameState.RoomStatus = roommodel.Ended
		manager.Unlock()

		// Send game end event
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
	currentQuestion := gameState.Questions.QuestionData[gameState.CurrentQuestionIndex]
	timeLimit := time.Duration(gameState.Questions.TimeLimit) * time.Minute

	// Store current question index for the goroutine
	currentIndex := gameState.CurrentQuestionIndex

	manager.Unlock()

	questEvent := questionEvent{
		QuestionIndex:  currentIndex + 1,
		TotalQuestions: len(gameState.Questions.QuestionData),
		Question:       currentQuestion,
		StartTime:      time.Now(),
		TimeLimit:      gameState.Questions.TimeLimit,
	}

	// Prepare question event
	questionData, err := json.Marshal(questEvent)
	if err != nil {
		l.Sugar().Error("json marshal failed", err)
		return err
	}
	questionEvent := Event{Type: EventNewQuestion, Payload: questionData}

	// Broadcast to all clients
	for client := range manager.clients[roomCode] {
		if !client.isBot {
			client.egress <- questionEvent
		}
	}

	// Notify bots about new question
	manager.broadcastToBots(ctx, roomCode, questionEvent)

	// Schedule next question after a delay
	go func() {
		l.Sugar().Infof("Question %d timer started for %v seconds", currentIndex+1, timeLimit.Seconds())
		timer := time.NewTimer(timeLimit)
		<-timer.C
		l.Sugar().Infof("Question %d timer completed", currentIndex+1)

		// Move to next question
		manager.Lock()
		if gameState, exists := manager.gameStates[roomCode]; exists {
			// Only increment if we're still on the same question
			// This prevents race conditions if something else modified the index
			if gameState.CurrentQuestionIndex == currentIndex {
				gameState.CurrentQuestionIndex++
				manager.Unlock()
				err = sendNextQuestion(ctx, manager, roomCode)
				if err != nil {
					l.Sugar().Error("send next question failed", err)
				}
			} else {
				manager.Unlock()
			}
		} else {
			manager.Unlock()
		}
	}()

	return nil
}

// Handle answer submissions
func SubmitAnswerHandler(ctx context.Context, event Event, c *Client) error {
	l := logs.GetLoggerctx(ctx)

	var submission quizmodel.AnswerReq
	if err := json.Unmarshal(event.Payload, &submission); err != nil {
		l.Sugar().Error("bad payload", err)
		return fmt.Errorf("bad payload: %v", err)
	}

	// users dont send their userID but bots send. so for users we fetch from context
	submission.UserID = c.userID

	// Set timestamp if not provided
	if submission.AnswerTime.IsZero() {
		submission.AnswerTime = time.Now()
	}

	// Get the game state
	c.manager.Lock()
	gameState, exists := c.manager.gameStates[c.roomCode]
	if !exists {
		c.manager.Unlock()
		l.Sugar().Error(fmt.Sprintf("game state not found for room %s", c.roomCode))
		return fmt.Errorf("game state not found for room %s", c.roomCode)
	}

	// Only process if game is in progress
	if gameState.RoomStatus != roommodel.Started || gameState.CurrentQuestionIndex >= len(gameState.Questions.QuestionData) {
		c.manager.Unlock()
		l.Sugar().Error("game is not in active question phase")
		return fmt.Errorf("game is not in active question phase")
	}

	//  TODO: getting the user data, updating the answer data, everything should happen in db

	// Get current question
	currentQuestion := gameState.Questions.QuestionData[gameState.CurrentQuestionIndex]

	// Check if answer is correct
	isCorrect := int(submission.AnswerOption) == currentQuestion.Answer

	// Update participant score
	found := false
	for i, participant := range gameState.Participants {
		if participant.UserID == c.userID {
			if isCorrect {
				// Calculate score based on answer speed
				answerTime := submission.AnswerTime.Sub(gameState.StartTime)
				speedBonus := float64(gameState.Questions.TimeLimit) - answerTime.Seconds()
				if speedBonus < 0 {
					speedBonus = 0
				}

				// Add score (base 100 + speed bonus up to 100)
				gameState.Participants[i].Score += 100 + int(speedBonus*3.33)
			}
			gameState.Participants[i].LastAnsweredQestion = currentQuestion.ID
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

		gameState.Participants = append(gameState.Participants, quizmodel.Participant{
			UserID:              c.userID,
			Username:            username,
			IsBot:               c.isBot,
			Score:               score,
			IsReady:             true,
			LastAnsweredQestion: currentQuestion.ID,
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

// NextQuestionHandler is manually called by the user by clicking the next question button
func NextQuestionHandler(ctx context.Context, event Event, c *Client) error {

	// get the room code
	l := logs.GetLoggerctx(ctx)

	// check the db if all users in the game have submitted the answers
	// roomMembers, err := room.ListRoomMembersByRoomCode(ctx, roommodel.RoomCodeReq{
	// 	UserID:   c.userID,
	// 	RoomCode: c.roomCode,
	// })
	// if err != nil {
	// 	return err
	// }

	// if the data comes from db
	// answers, err := quiz.ListAnswersByRoomCode(ctx, c.roomCode)
	// if err != nil {
	// 	return fmt.Errorf("list Answers by room code failed", err)
	// }
	// isAllMembersSubmitted := true
	// answeredMap := make(map[uuid.UUID]bool)
	// for _, answer := range answers {
	// 	answeredMap[answer.UserID] = true
	// }
	// for _, roomMember := range roomMembers {
	// 	if !roomMember.IsBot {
	// 		if _, ok := answeredMap[roomMember.UserID]; !ok {
	// 			isAllMembersSubmitted = false
	// 			break
	// 		}
	// 	}
	// }

	// go in-memory
	c.manager.Lock()
	gameState, exists := c.manager.gameStates[c.roomCode]
	c.manager.Unlock()
	if !exists {
		l.Sugar().Error(fmt.Sprintf("game state not found for room %s", c.roomCode))
		return fmt.Errorf("game state not found for room %s", c.roomCode)
	}
	isAllMembersSubmitted := true
	currentQuestion := gameState.Questions.QuestionData[gameState.CurrentQuestionIndex]

	for _, participant := range gameState.Participants { // TODO: This logic is not the best logic here we are checking if all of them have answeredby checking if all have the same latest quest id
		if !participant.IsBot && currentQuestion.ID != participant.LastAnsweredQestion {
			isAllMembersSubmitted = false
			break

		}
	}
	// if yes
	if isAllMembersSubmitted && len(gameState.Participants) > 0 {
		// move to next question
		c.manager.Lock()
		if gameState, exists := c.manager.gameStates[c.roomCode]; exists {
			// Only increment if we're still on the same question
			// This prevents race conditions if something else modified the index

			gameState.CurrentQuestionIndex++
			c.manager.Unlock()
			return sendNextQuestion(ctx, c.manager, c.roomCode)
		} else {
			c.manager.Unlock()
		}
	} else {
		// TODO: if it reaches this error connection gets closed. instead we just need to print the error in UI
		l.Sugar().Error("Cannot proceed to the next question until all users have submitted their answers or the time limit has expired.")
		em := quizmodel.QuizError{
			Message: "Cannot proceed to the next question until all users have submitted their answers or the time limit has expired.",
		}
		errorData, _ := json.Marshal(em)
		ackEvent := Event{Type: "game_error", Payload: errorData}

		// Send acknowledgment only to the client who submitted
		c.egress <- ackEvent

	}

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

	// TODO: this api is wrong
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
	roomMembers, err := room.ListRoomMembersByRoomCode(ctx, roommodel.RoomCodeReq{
		RoomCode: c.roomCode,
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
			TotalRounds:  totalRounds,
			Questions:    &quizmodel.Question{},
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
	readLimit := viper.GetInt64("ws.maxReadSize")
	if readLimit <= 0 {
		readLimit = int64(maxReadLimit) // Use default if not configured
	}
	c.connection.SetReadLimit(readLimit) // max read size limit bytes

	for {
		_, payload, err := c.connection.ReadMessage()
		if err != nil {
			// Check specific error types
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
				websocket.CloseNoStatusReceived) {
				l.Sugar().Errorf("unexpected close error: %v", err)
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				l.Sugar().Errorf("read timeout: %v", err)
			} else {
				l.Sugar().Errorf("error reading message: %v", err)
			}
			break
		}
		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			l.Sugar().Error("error unmarshalling message:", err)
			break
		}
		if err := c.manager.routeEvent(ctx, request, c); err != nil {
			l.Sugar().Error("error routing event:", err)
			// send the error received if any and display it in ui
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
			err := c.connection.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				l.Sugar().Error(err)
			}
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
