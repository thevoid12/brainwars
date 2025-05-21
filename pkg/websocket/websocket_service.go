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
	usermodel "brainwars/pkg/users/model"
	"brainwars/pkg/util"
	"brainwars/web/ui/handlers"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

var websocketUpgrader = websocket.Upgrader{
	// 	CheckOrigin:     checkOrigin, TODO: SETUP ORIGIN CHECK
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

var ErrEventNotSupported = errors.New("this event type is not supported")

func checkOrigin(r *http.Request) bool {
	return r.Header.Get("Origin") == "https://localhost:8080" // TODO: move it to viper config
}

type Manager struct {
	clients    map[string]ClientList            // map key is roomCode value is the list of clients in that room
	botClients map[string]map[uuid.UUID]*Client // Separate map for bots per room map key is roomid, map[roomid]map[botid]client

	sync.RWMutex
	handlers   map[string]EventHandler
	roomStates map[string]*roommodel.RoomStatus
	gameStates map[string]*quizmodel.GameState
}

type ClientList map[*Client]bool
type Client struct {
	QuestionCancel context.CancelFunc // cancel func for the current question's goroutine
	TOC            time.Time          // Time of creation
	connection     *websocket.Conn
	manager        *Manager
	egress         chan Event
	botEvents      chan Event // Only used for bot clients
	roomCode       string
	isBot          bool              // Flag to identify bot clients
	botType        usermodel.BotType // Empty for real users, "30sec", "1min", "2min" for bots
	userID         uuid.UUID         // Store the user ID for easier reference
	UserName       string
	UserStatus     usermodel.UserStatus
	room           *roommodel.Room
	ansHistory     map[uuid.UUID]map[uuid.UUID]*quizmodel.AnswerReq // map[questionD]map[userID]answerIW
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
		botClients: make(map[string]map[uuid.UUID]*Client),

		roomStates: make(map[string]*roommodel.RoomStatus),
		gameStates: make(map[string]*quizmodel.GameState),
	}
	m.setupEventHandlers()
	m.MemoryCleanup(ctx)
	go m.startClientHealthCheck(ctx)
	return m
}

func NewClient(conn *websocket.Conn, manager *Manager, roomCode string, isBot bool, botType usermodel.BotType, userID uuid.UUID, userName string, room *model.Room) *Client {

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
		ansHistory: make(map[uuid.UUID]map[uuid.UUID]*quizmodel.AnswerReq),
		TOC:        time.Now(),
		UserStatus: usermodel.UserReady,
		UserName:   userName,
	}
}

// TODO: here lock is open for a long time. find a way to reduce lock time
func (m *Manager) MemoryCleanup(ctx context.Context) {
	l := logs.GetLoggerctx(ctx)
	ticker := time.NewTicker(time.Minute * time.Duration(viper.GetInt("cacheCleaner.repeatIntervalMinutes")))
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				m.Lock()
				interval := time.Minute * time.Duration(viper.GetInt("cacheCleaner.intervalMinutes"))
				for roomCode, clients := range m.clients {
					for client := range clients {
						fmt.Println(client.TOC.Format("2006-01-02 15:04:05"))
						expiry := client.TOC.Add(interval)
						fmt.Println(expiry.Format("2006-01-02 15:04:05"))
						fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
						// Check if the client is inactive (e.g., not sending/receiving messages)
						if time.Now().After(expiry) {
							// if client is not closed then close it TODO: what will happen if  the connection is already closed?
							if client.connection != nil {
								client.connection.Close()
							}
							// Remove the client from the list
							delete(clients, client)
							l.Sugar().Info("Client %s removed from room %s due to inactivity", client.userID.String(), roomCode)
						}
					}
					// If no clients are left in the room, remove the room from the manager
					if len(clients) == 0 {
						delete(m.clients, roomCode)
					}
				}

				for roomCode, clients := range m.botClients {
					for botid, client := range clients {
						fmt.Println(client.TOC.Format("2006-01-02 15:04:05"))
						expiry := client.TOC.Add(interval)
						fmt.Println(expiry.Format("2006-01-02 15:04:05"))
						fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
						// Check if the client is inactive (e.g., not sending/receiving messages)
						if time.Now().After(expiry) {
							// if client is not closed then close it TODO: what will happen if  the connection is already closed?
							if client.connection != nil {
								client.connection.Close()
							}
							// Remove the client from the list
							delete(clients, botid)
							l.Sugar().Info("bot Client %s removed from room %s due to inactivity", client.userID.String(), roomCode)
						}
					}
					// If no clients are left in the room, remove the room from the manager
					if len(clients) == 0 {
						delete(m.clients, roomCode)
					}
				}
				m.Unlock()
			}
		}
	}()
	// Stop the ticker when the context is done
	go func() {
		<-ctx.Done()
		ticker.Stop()
		done <- true
	}()
}

// when user joins a room this serveWs handler is called
// TODO: render error templates
func (m *Manager) ServeWS(c *gin.Context) {
	ctx := c.Request.Context()
	l := logs.GetLoggerctx(ctx)
	userInfo := util.GetUserInfoFromctx(ctx)
	userID := userInfo.ID

	// get the roomID from the query params
	roomCode := c.Query("roomCode")
	if roomCode == "" {
		l.Sugar().Error("room code not found", nil)
		return
	}

	// Check if the user is already in the room so when he refreshes the page
	// he is pushed out of the page and the connection is closed since the game is realtime multiplayer
	m.Lock()
	gameState, exists := m.gameStates[roomCode]
	m.Unlock()

	if exists {
		if gameState.RoomStatus == roommodel.Started {
			l.Sugar().Error("user already in the room (page refreshed)")
			//http.Error(c.Writer, "REFRESHED_PAGE", http.StatusForbidden)
			//c.Redirect(http.StatusFound, "/bw/home/")
			handlers.RenderErrorTemplate(c, "home.html", "refreshing page kicks you out of the game since the game runs realtime!", nil)
			return
		}
	} else {
		questions, err := quiz.ListQuestionsByRoomCode(ctx, roomCode)
		if err != nil {
			l.Sugar().Error("list questions by room code failed", err)
			return
		}

		totalQuestions := questions.QuestionCount
		// Check if the room needs to be initialized
		m.initializeRoomGameState(ctx, roomCode, totalQuestions)
	}

	conn, err := websocketUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		l.Sugar().Error("websocket upgrade error:", err)
		return
	}

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
		l.Sugar().Error("there are no room members")
		return
	}

	client := NewClient(conn, m, roomCode, false, "", userID, roomMember.UserDetails.UserName, roomDetails)
	m.addClient(client)
	go m.readMessages(ctx, client)
	go m.writeUsersMessages(ctx, client)

	// checking !exists here because we need to initialize bots just once per room that is for the first time alone
	if !exists {
		// When a human player joins, set bots to ready state
		// even if 1 user joins the room then we instentaniously set up all the bots to ready state for the game to start.
		// we get the list of bots from list all members in a room where bots are members as ready
		m.setupBotsForRoom(ctx, roomCode, roomDetails)
	}

	// if the game is a single player game since the user is ready and bots are ready as well
	//  we automatically display the first question. in terms of multiplayer game a button needs to be triggered
	// to start the game
	if roomDetails.GameType == roommodel.SP {
		err = m.sendRoomMemberState(ctx, roomCode, client, usermodel.UserReady)
		if err != nil {
			l.Sugar().Error("send single player user ready state failed ", err)
			return
		}
		err = StartGameMessageHandler(ctx, Event{}, client)
		if err != nil {
			return
		}

	} else { // if it is a multiplayer the user first just joins the common game lobby. he clicks a button and makes himself ready later
		err = m.sendRoomMemberState(ctx, roomCode, client, usermodel.UserJoined)
		if err != nil {
			l.Sugar().Error("setup user for room failerd ", err)
			return
		}
	}

}

// send roomMemberState helps to send user and bots state to the ui in lobby
func (m *Manager) sendRoomMemberState(ctx context.Context, roomCode string, c *Client, state usermodel.UserStatus) error {
	l := logs.GetLoggerctx(ctx)
	// Use context with timeout for database operation
	// dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancel()

	// var userDetails *usermodel.UserInfo
	var err error
	// userDetails = util.GetUserInfoFromctx(ctx)

	m.Lock()
	_, exists := m.clients[roomCode][c]
	if exists {
		delete(m.clients[roomCode], c)

		c.UserStatus = state
		m.clients[roomCode][c] = true
	}
	botclient := m.botClients[roomCode][c.userID]
	if botclient != nil {
		m.botClients[roomCode][c.userID].UserStatus = state // always bot ready
	}

	userStateNotification := []Payload{}
	for _, botclient := range m.botClients[roomCode] {
		userStateNotification = append(userStateNotification, Payload{UserName: botclient.UserName,
			Data: string(botclient.UserStatus),
		})
	}
	for client := range m.clients[roomCode] {
		userStateNotification = append(userStateNotification, Payload{UserName: client.UserName,
			Data: string(client.UserStatus),
		})
	}
	//TODO: we can add time and sort by time for uniform list all the time
	m.Unlock()

	// Retry logic for getting user details
	// for retries := 0; retries < 3; retries++ {
	// 	userDetails, err = user.GetUserDetailsbyID(ctx, userID)
	// 	if err == nil {
	// 		break
	// 	}
	// 	l.Sugar().Warnf("Get user details attempt %d failed: %v. Retrying...", retries+1, err)
	// 	time.Sleep(1 * time.Second)
	// }
	// Set the user to ready state
	// Notify all clients that this user is ready

	data, err := json.Marshal(userStateNotification)
	if err != nil {
		l.Sugar().Error(" user state room notification json marshal failed", err)
		return err
	}

	JoinEvent := Event{Type: EventLobbyState, Payload: data}

	// TODO: update user room member's state to ready
	// for retries := 0; retries < 3; retries++ {
	// 	userDetails, err = room.UpdateRoomMemberByID(ctx)
	// 	if err == nil {
	// 		break
	// 	}
	// 	l.Sugar().Warnf("Get user details attempt %d failed: %v. Retrying...", retries+1, err)
	// 	time.Sleep(1 * time.Second)
	// }

	m.Lock()
	clientList := m.clients[roomCode]
	m.Unlock()
	// Broadcast to all user clients in the room
	for client := range clientList {
		if !client.isBot && client.connection != nil {
			client.egress <- JoinEvent
		}
	}

	return nil
}

// // Set up user to be in game ready state when a human player clicks ready
// func (m *Manager) sendUserReadyState(ctx context.Context, roomCode string) error {
// 	l := logs.GetLoggerctx(ctx)

// 	var userDetails *usermodel.UserInfo
// 	var err error
// 	userDetails = util.GetUserInfoFromctx(ctx)

// 	joinNotification := Payload{
// 		UserName: userDetails.UserName,
// 		Data:     fmt.Sprintf("User %s ready", userDetails.UserName),
// 		Time:     time.Now(),
// 	}

// 	data, err := json.Marshal(joinNotification)
// 	if err != nil {
// 		l.Sugar().Error(" user joined room notification json marshal failed", err)
// 		return err
// 	}

// 	JoinEvent := Event{Type: EventReadyGame, Payload: data}

// 	// TODO: update user room member's state to ready
// 	// for retries := 0; retries < 3; retries++ {
// 	// 	userDetails, err = room.UpdateRoomMemberByID(ctx)
// 	// 	if err == nil {
// 	// 		break
// 	// 	}
// 	// 	l.Sugar().Warnf("Get user details attempt %d failed: %v. Retrying...", retries+1, err)
// 	// 	time.Sleep(1 * time.Second)
// 	// }

// 	m.Lock()
// 	clientList := m.clients[roomCode]
// 	m.Unlock()
// 	// Broadcast to all user clients in the room
// 	for client := range clientList {
// 		if !client.isBot {
// 			client.egress <- JoinEvent
// 		}
// 	}

// 	return nil
// }

// Start game handler
func StartGameMessageHandler(ctx context.Context, event Event, c *Client) error {
	l := logs.GetLoggerctx(ctx)

	// em := quizmodel.QuizError{
	// 	Message: "this is a test error",
	// }
	// errorData, _ := json.Marshal(em)
	// ackEvent := Event{Type: EventGameError, Payload: errorData}

	// // Send acknowledgment only to the client who submitted
	// c.egress <- ackEvent

	// Get the game state for this room
	c.manager.Lock()
	gameState, exists := c.manager.gameStates[c.roomCode]
	c.manager.Unlock()
	if !exists {
		return fmt.Errorf("game state not found for room %s", c.roomCode)
	}

	if gameState.Questions.QuestionData == nil { // init setup all questions for that room
		c.manager.Lock()
		// Update game state
		gameState.RoomStatus = roommodel.Started
		gameState.StartTime = time.Now()
		gameState.CurrentQuestionIndex = 0
		c.manager.Unlock()

		// Fetch questions from the database
		questions, err := quiz.ListQuestionsByRoomCode(ctx, c.roomCode) // TODO: instead of storing all generated questions in the db and fetching them here we need to store in memory and use it and slowly update it back to the database
		if err != nil {
			l.Sugar().Error("failed to fetch questions:", err)
			return err
		}

		// Store questions in game state
		c.manager.Lock()
		gameState.Questions = questions
		c.manager.Unlock()
	}
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

		manager.Lock()
		clients := manager.clients[roomCode]
		manager.Unlock()
		// Broadcast to all clients
		for client := range clients {
			if !client.isBot {
				client.egress <- endEvent
			}
		}

		// Notify bots about game end so that it can update its answer history and cleanup
		botGameoverEvent := Event{
			Type:    EventBotGameOver,
			Payload: json.RawMessage{},
		}
		manager.broadcastToBots(ctx, roomCode, botGameoverEvent)

		// manager.broadcastToBots(ctx, roomCode, endEvent) TODO: somehow notify bots to exit the routine clear the client memory
		//	quiz.HandleLastQuestion(ctx,roomCode,)
		events := []Event{endEvent}
		eventjson, err := json.Marshal(events)
		if err != nil {
			l.Sugar().Error("json marshal failed", err)
			return err
		}
		err = room.UpdateRoomMetaAndStatus(ctx, roommodel.RoomMetaReq{
			RoomCode: roomCode,
			RoomMeta: string(eventjson)})
		if err != nil {
			return err
		}

		// updating leader board
		for _, player := range gameState.Participants {
			err := room.UpdateLeaderBoard(ctx, &roommodel.EditLeaderBoardReq{
				RoomCode: roomCode,
				UserID:   player.UserID,
				Score:    float64(player.Score),
			})
			if err != nil {
				return err
			}
		}
		//TODO: find a better way to update the database (like a queue kind of thingy to
		// update db later and clear the memory as well after updating the pgsql db
		// updating the answer history in answer table
		manager.Lock()
		clients = manager.clients[roomCode]
		manager.Unlock()
		for client := range clients {
			err = updateAnswerHistory(ctx, client.ansHistory)
			if err != nil {
				return err
			}
		}

		return nil
	}

	// Get the current question
	currentQuestion := gameState.Questions.QuestionData[gameState.CurrentQuestionIndex]
	timeLimit := time.Duration(gameState.Questions.TimeLimit) * time.Minute
	currentQuestion.Answer = -1 // making sure that the answer isnt shown in ws
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
	c.manager.Lock() // TODO: check if we are locking for too long
	gameState, exists := c.manager.gameStates[c.roomCode]
	if !exists {
		c.manager.Unlock()
		l.Sugar().Error(fmt.Sprintf("game state not found for room %s", c.roomCode))
		return fmt.Errorf("game state not found for room %s", c.roomCode)
	}

	// Only process if game is in progress
	if gameState.RoomStatus != roommodel.Started || gameState.CurrentQuestionIndex >= len(gameState.Questions.QuestionData) {
		// we reach here after game over
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
			found = true
			// check to make sure that no scores to be calculated when the same option is clicked again and again
			if gameState.Participants[i].LastChoosenOption == int(submission.AnswerOption) && gameState.Participants[i].LastAnsweredQestion == currentQuestion.ID {
				break
			}
			// if he has already answered the question we will -50 the points
			if participant.LastAnsweredQestion == currentQuestion.ID {
				gameState.Participants[i].Score -= 50
			}
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
			gameState.Participants[i].LastChoosenOption = int(submission.AnswerOption)

			// Update the answer history
			if _, exists := c.ansHistory[currentQuestion.ID]; !exists {
				c.ansHistory[currentQuestion.ID] = make(map[uuid.UUID]*quizmodel.AnswerReq)
			}
			c.ansHistory[currentQuestion.ID][c.userID] = &quizmodel.AnswerReq{
				RoomCode:       c.roomCode,
				UserID:         submission.UserID,
				QuestionID:     currentQuestion.ID,
				QuestionDataID: submission.QuestionDataID,
				AnswerOption:   submission.AnswerOption,
				IsCorrect:      isCorrect,
				AnswerTime:     submission.AnswerTime,
				CreatedBy:      "system",
			}
			break
		}
	}

	// If participant not found, add them
	if !found {
		// Get user info from database
		var username string
		// TODO: Replace with actual database query to get username
		if c.isBot {
			username = fmt.Sprintf("Bot-%s", c.botType)
		} else {
			userInfo := util.GetUserInfoFromctx(ctx)
			username = userInfo.UserName
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
			LastChoosenOption:   int(submission.AnswerOption),
			IsExited:            false,
		})
		// Update the answer history
		if _, exists := c.ansHistory[currentQuestion.ID]; !exists {
			c.ansHistory[currentQuestion.ID] = make(map[uuid.UUID]*quizmodel.AnswerReq)
		}
		c.ansHistory[currentQuestion.ID][c.userID] = &quizmodel.AnswerReq{
			RoomCode:       c.roomCode,
			UserID:         submission.UserID,
			QuestionID:     currentQuestion.ID,
			QuestionDataID: submission.QuestionDataID,
			AnswerOption:   submission.AnswerOption,
			IsCorrect:      isCorrect,
			AnswerTime:     submission.AnswerTime,
			CreatedBy:      "system",
		}
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
	// update the db
	err := sendLiveLeaderBoard(ctx, c.manager, c.roomCode)
	if err != nil {
		return err
	}

	// Send acknowledgment only to the client who submitted
	c.egress <- ackEvent
	return nil
}

// NextQuestionHandler is manually called by the user by clicking the next question button
func NextQuestionHandler(ctx context.Context, event Event, c *Client) error {

	// get the room code
	l := logs.GetLoggerctx(ctx)

	// check the db if all users in the game have submitted the answers
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
		if !participant.IsBot && !participant.IsExited && currentQuestion.ID != participant.LastAnsweredQestion {
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
		ackEvent := Event{Type: EventGameError, Payload: errorData}

		// Send acknowledgment only to the client who submitted
		c.egress <- ackEvent

	}

	return nil
}

func updateAnswerHistory(ctx context.Context, ansHistory map[uuid.UUID]map[uuid.UUID]*quizmodel.AnswerReq) error {
	l := logs.GetLoggerctx(ctx)
	for _, answermap := range ansHistory {
		for _, answerreq := range answermap {
			err := quiz.CreateAnswer(ctx, answerreq)
			if err != nil {
				l.Sugar().Error("create answer failed", err)
				return err // todo:retry logic needs to be added
			}
		}
	}
	return nil
}

func sendLiveLeaderBoard(ctx context.Context, manager *Manager, roomCode string) error {
	l := logs.GetLoggerctx(ctx)

	manager.Lock()
	gameState, exists := manager.gameStates[roomCode]
	manager.Unlock()
	if !exists {
		l.Sugar().Error("game state not found for room %s", roomCode)
		return fmt.Errorf("game state not found for room %s", roomCode)
	}

	sort.Slice(gameState.Participants, func(i, j int) bool {
		return gameState.Participants[i].Score > gameState.Participants[j].Score
	})

	// Send game end event
	lbPayload := struct {
		Message string                  `json:"message"`
		Scores  []quizmodel.Participant `json:"scores"`
	}{
		Message: "The live leaderboard is updated.",
		Scores:  gameState.Participants,
	}

	lbData, _ := json.Marshal(lbPayload)
	lbEvent := Event{Type: EventLeaderBoard, Payload: lbData}

	// Broadcast to all clients including bots
	for client := range manager.clients[roomCode] {
		if !client.isBot {
			client.egress <- lbEvent
		}
	}

	// Also notify bots about game end
	// manager.broadcastToBots(ctx, roomCode, endEvent) TODO: somehow notify bots to exit the routine clear the client memory
	//	quiz.HandleLastQuestion(ctx,roomCode,)

	return nil
}

// Modified ReadyGameMessageHandler to check if all participants are ready and start game
func ReadyGameMessageHandler(ctx context.Context, event Event, c *Client) error {
	l := logs.GetLoggerctx(ctx)

	userDetails := util.GetUserInfoFromctx(ctx)

	// TODO: this api is wrong we can update inmemory

	// Update the room member's ready status
	err := room.UpdateRoomMemberStatusByRoomCodeAndUserID(ctx, &roommodel.RoomCodeReq{
		UserID:   userDetails.ID,
		RoomCode: c.roomCode,
	}, roommodel.ReadyQuiz)
	if err != nil {
		l.Sugar().Error("Failed to update ready status", err)
		return err
	}

	// Check if all room members are ready TODO: we dont have to go to database room member details are in memory
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
		if member.RoomMemberStatus != roommodel.ReadyQuiz && !member.IsBot {
			allReady = false
			break
		}
	}

	// Everyone is ready, start the game
	gameReadyNotification := struct {
		UserName string    `json:"username"`
		Message  string    `json:"message"`
		Time     time.Time `json:"time"`
	}{
		UserName: userDetails.UserName,
		Message:  fmt.Sprintf("user %s is ready", userDetails.UserName),
		Time:     time.Now(),
	}

	readyEventJson, err := json.Marshal(gameReadyNotification)
	if err != nil {
		l.Sugar().Error("ready event json marshal failed", err)
	}
	readyEvent := Event{Type: EventReadyGame, Payload: readyEventJson}

	// Broadcast start notification to all clients
	c.manager.Lock()
	clients := c.manager.clients[c.roomCode]
	c.manager.Unlock()
	for client := range clients {
		if client.isBot {
			continue
		}
		client.egress <- readyEvent
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

		c.manager.Lock()
		clients := c.manager.clients[c.roomCode]
		c.manager.Unlock()

		// Broadcast start notification to all clients
		for client := range clients {
			if client.isBot {
				continue
			}
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

// leave the game room between game
func LeaveGameRoomHandler(ctx context.Context, event Event, c *Client) error {
	// l := logs.GetLoggerctx(ctx)
	userDetails := util.GetUserInfoFromctx(ctx)

	// before leaving remove this client
	leaveGameNotfication := struct {
		UserName string    `json:"username"`
		Message  string    `json:"message"`
		Time     time.Time `json:"time"`
	}{
		UserName: userDetails.UserName,
		Message:  fmt.Sprintf("user %s left the room", userDetails.UserName),
		Time:     time.Now(),
	}

	startData, _ := json.Marshal(leaveGameNotfication)
	eventPayload := Event{Type: EventLeaveRoom, Payload: startData}
	// todo: remove that client from the manager

	// Broadcast leave notification to that client

	// for client := range clients {
	// 	if !client.isBot {
	// 		client.egress <- eventPayload
	// 	}
	// }
	c.egress <- eventPayload

	c.manager.Lock()
	// clients := c.manager.clients[c.roomCode]
	gamestate := c.manager.gameStates[c.roomCode]
	c.manager.Unlock()
	// newclientList := make(ClientList)
	// for client := range clients {
	// 	if client.userID == c.userID {
	// 		continue
	// 	}
	// 	newclientList[client] = true
	// }

	newParticipants := []quizmodel.Participant{}

	for _, p := range gamestate.Participants {
		if p.UserID == c.userID {
			p.IsExited = true
		}
		newParticipants = append(newParticipants, p)
	}

	c.manager.Lock()
	// c.manager.clients[c.roomCode] = newclientList
	c.manager.gameStates[c.roomCode].Participants = newParticipants
	c.manager.Unlock()

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
		// client.connection.Close()
	}
	m.Unlock()
}

func (m *Manager) addBot(roomCode string, bot *Client) {
	m.Lock()
	defer m.Unlock()

	if m.botClients == nil {
		m.botClients = make(map[string]map[uuid.UUID]*Client)
	}

	if _, exists := m.botClients[roomCode]; !exists {
		m.botClients[roomCode] = make(map[uuid.UUID]*Client)
	}
	m.botClients[roomCode][bot.userID] = bot
}

func (m *Manager) removeBot(bot *Client) {
	m.Lock()
	defer m.Unlock()
	if _, exists := m.botClients[bot.roomCode]; exists {
		delete(m.botClients[bot.roomCode], bot.userID)
		if len(m.botClients[bot.roomCode]) == 0 {
			delete(m.botClients, bot.roomCode)
		}
	}
}

// Add this to your manager to periodically check clients
func (m *Manager) startClientHealthCheck(ctx context.Context) {
	// l := logs.GetLoggerctx(ctx)

	// ticker := time.NewTicker(30 * time.Second)
	// defer ticker.Stop()

	// for {
	// 	select {
	// 	case <-ticker.C:
	// 		m.RLock()
	// 		for roomCode, clients := range m.clients {
	// 			log.Printf("Room %s has %d clients", roomCode, len(clients))
	// 			for client := range clients {
	// 				if client.isBot {
	// 					continue
	// 				}
	// 				startEvent := Event{Type: "test_event"}
	// 				se, err := json.Marshal(startEvent)
	// 				if err != nil {
	// 					l.Sugar().Error(err)
	// 				}
	// 				err = client.connection.WriteMessage(websocket.TextMessage, se)
	// 				if err != nil {
	// 					l.Sugar().Error("connection closed for the client(cannot write the user id)", client.userID)
	// 				}
	// 				log.Printf("Client %s in room %s is connected", client.userID, roomCode)
	// 			}
	// 		}
	// 		m.RUnlock()
	// 	}
	// }
}

func (m *Manager) readMessages(ctx context.Context, c *Client) {
	l := logs.GetLoggerctx(ctx)
	l.Sugar().Infof("Read goroutine started for user %s in room %s", c.userID, c.roomCode)

	defer func() {
		l.Sugar().Infof("Ending read goroutine for user %s", c.userID)
		c.manager.removeClient(c)
	}()

	// I am registering pong handler here so that gorilla mux can handle pong
	// internally whenever i ping is written to the ws
	c.connection.SetPongHandler(c.pongHandler)

	readLimit := viper.GetInt64("ws.maxReadSize")
	if readLimit <= 0 {
		readLimit = int64(maxReadLimit)
	}
	c.connection.SetReadLimit(readLimit)

	for {
		select {
		case <-ctx.Done():
			l.Sugar().Infof("Context cancelled for readMessages, user %s", c.userID)
			return
		default:
			if c.connection == nil { // Should not happen for human clients after successful connection
				l.Sugar().Errorf("User %s has nil connection in readMessages loop.", c.userID)
				return
			}
			_, payload, err := c.connection.ReadMessage()

			if err != nil {
				l.Sugar().Errorf("Error reading message for user %s: %v. Connection type: %T", c.userID, err, c.connection)

				if websocket.IsUnexpectedCloseError(err,
					websocket.CloseGoingAway,
					websocket.CloseNormalClosure,
					websocket.CloseNoStatusReceived) {
					l.Sugar().Errorf("unexpected close error: %v", err)
				} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					l.Sugar().Errorf("read timeout: %v", err)
				}

				return
			}
			l.Sugar().Debugf("Raw message received from user %s: payload %s", c.userID, string(payload))

			var request Event
			if err := json.Unmarshal(payload, &request); err != nil {
				l.Sugar().Error("error unmarshalling message:", err)
				return
			}

			// Add additional logging here to debug the event routing
			l.Sugar().Infof("Received event type %s from user %s", request.Type, c.userID)

			if err := m.routeEvent(ctx, request, c); err != nil {
				l.Sugar().Error("error routing event:", err)
			}
		}
	}
}

func (m *Manager) writeUsersMessages(ctx context.Context, c *Client) {
	l := logs.GetLoggerctx(ctx)

	l.Sugar().Info("Client connected for write user messages")
	ticker := time.NewTicker(pingInterval)

	defer func() {
		ticker.Stop()
		m.removeClient(c)
	}()
	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				c.connection.WriteMessage(websocket.CloseMessage, nil)
				l.Info("connection cloed since not okay")
				return
			}
			data, err := json.Marshal(message)
			if err != nil {
				l.Sugar().Error("json message marshall failed", err)
			}

			l.Sugar().Info(fmt.Sprintf("the data received by the writer is %s, for the user %s", string(data), c.userID.String()))
			err = c.connection.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				l.Sugar().Error(err)
			}
		case <-ticker.C:
			log.Println("ping")
			l.Sugar().Debugf("Sending ping to user %s", c.userID)
			err := c.connection.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				l.Sugar().Error("ping error", err)
				ticker.Stop()
				c.manager.removeClient(c)
				return
			}
		case <-ctx.Done():
			l.Info("Context canceled, stopping writer")
			ticker.Stop()
			return

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
